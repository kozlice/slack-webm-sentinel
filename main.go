package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/caarlos0/env"
	"github.com/google/uuid"
	"github.com/nlopes/slack"
	log "github.com/sirupsen/logrus"
	"github.com/xfrr/goffmpeg/transcoder"
)

type NotifyMode string

const (
	Reaction NotifyMode = "reaction"
	Message  NotifyMode = "message"
)

type config struct {
	SlackApiToken string     `env:"SLACK_API_TOKEN,required"`
	NotifyMode    NotifyMode `env:"NOTIFY_MODE" envDefault:"reaction"`
	TempDir       string     `env:"TEMP_DIR" envDefault:"/tmp" envExpand:"true"`
	Debug         bool       `env:"DEBUG" envDefault:"false"`
}

var api *slack.Client
var rtm *slack.RTM
var webmUrlMatcher = regexp.MustCompile(`https?://\S+\.webm`)
var logger = log.New()
var currentBotId = ""
var cfg = config{}

func handle(url string, ev *slack.MessageEvent) {
	logger.Debug(fmt.Sprintf("Found matching URL %s", url))

	switch cfg.NotifyMode {
	case Reaction:
		reaction(ev, "ok_hand")
	case Message:
		text := slack.MsgOptionText(fmt.Sprintf("Found an URL with .webm: %s\nWill try to make .mp4", url), false)
		api.SendMessage(ev.Channel, text)
	default:
	}

	// Download source .webm
	logger.Debug(fmt.Sprintf("Downloading %s", url))
	source, err := download(url)
	if err != nil {
		logger.Error(fmt.Sprintf("Could not download %s, error: %v", url, err))
		return
	}
	logger.Debug(fmt.Sprintf("Successfully downloaded %s into %s", url, source))

	// Convert .webm to .mp4
	logger.Debug(fmt.Sprintf("Converting %s into .mp4", source))
	result, err := convert(source)
	// Remove source file regardless of success
	os.Remove(source)
	if err != nil {
		logger.Error(fmt.Sprintf("Could not convert %s, error: %v", source, err))
		return
	}
	logger.Debug(fmt.Sprintf("Successfully encoded %s into .mp4", source))

	// Send message with .mp4
	logger.Debug(fmt.Sprintf("Uploading converted file %s", result))
	err = send(url, result, ev.Channel)
	// Remove converted file regardless of success
	os.Remove(result)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to upload .mp4 file, error: %v", err))
		return
	}
	logger.Debug(fmt.Sprintf("Uploaded %s", url))
}

func download(url string) (string, error) {
	filename := cfg.TempDir + "/" + uuid.New().String() + ".webm"

	out, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return filename, nil
}

func convert(source string) (string, error) {
	dst := strings.Replace(source, ".webm", ".mp4", -1)
	trans := new(transcoder.Transcoder)

	err := trans.Initialize(source, dst)

	if err != nil {
		return "", err
	}

	done := trans.Run(false)
	err = <-done

	return dst, nil
}

func send(url string, result string, channel string) error {
	f, err := os.Open(result)
	if err != nil {
		return err
	}

	_, err = api.UploadFile(slack.FileUploadParameters{
		Reader:   f,
		Filename: path.Base(url),
		Channels: []string{channel},
	})

	return err
}

func reaction(ev *slack.MessageEvent, emoji string) {
	itemRef := slack.ItemRef{
		Channel:   ev.Channel,
		Timestamp: ev.Timestamp,
	}
	api.AddReaction(emoji, itemRef)
}

func main() {
	// Configure logger
	logger.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})

	// Read config
	if err := env.Parse(&cfg); err != nil {
		logger.Error(fmt.Sprintf("Failed to parse config: %+v", err))
		os.Exit(1)
	}

	if cfg.Debug {
		logger.SetLevel(log.DebugLevel)
	}

	// Initialize API
	api = slack.New(cfg.SlackApiToken)

	// Find out and remember bot's ID
	auth, err := api.AuthTest()
	if err != nil {
		logger.Error(fmt.Sprintf("Could not complete auth test, error: %v", err))
		os.Exit(1)
	}

	user, err := api.GetUserInfo(auth.UserID)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to fetch info about bot's user, error: %v", err))
		os.Exit(1)
	}

	currentBotId = user.Profile.BotID

	// Fire up RTM
	rtm = api.NewRTM()
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.HelloEvent:
			logger.Debug("Connected to RTM")

		case *slack.MessageEvent:
			// Skip messages from self, otherwise endless loop will occur
			if ev.BotID == currentBotId {
				continue
			}

			// Find all URLs ending with `.webm` and handle them
			matches := webmUrlMatcher.FindAllStringSubmatch(ev.Text, -1)
			for _, match := range matches {
				go handle(match[0], ev)
			}

		default:
			// Ignore all other events
		}
	}
}
