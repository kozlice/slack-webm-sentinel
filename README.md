# Slack webm sentinel

Slack doesn't embed a preview for .webm files. However, it does for .mp4.
This bot tracks .webm links, downloads them, converts file to .mp4 and posts it back
to the channel.

Configuration env vars:
* `SLACK_API_TOKEN` - bot access token, required
* `NOTIFY_MODE` - can be one of
    * `message` - bot will post a message when it finds a URL ending with .webm
    * `reaction` - will add emoji to message with .webm links (default)
    * `none` - don't notify at all
* `DEBUG` - optional, will set logging level to debug
* `TEMP_DIR` - dir to store downloaded .webm files and .mp4 output

Run it in Docker:

```shell script
docker run -e SLACK_API_TOKEN=xoxb-... kozlice/slack-webm-sentinel
```

