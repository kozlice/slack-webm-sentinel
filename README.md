# Slack webm sentinel

Slack doesn't have an inline player for .webm links.

This bot fixes the problem: it tracks .webm links in messages, downloads them, converts into .mp4 and posts them back
into the channel.

### Configuration

Done via env variables:

|Name|Required?|Description|
|---|---|---|
|`SLACK_API_TOKEN`|Yes|Bot access token (starts with `xoxb-...`)|
|`NOTIFY_MODE`    |No |Can be `message` (bot will post a message when it finds URL), `reaction` (default, will add emoji to post with link) or `none`|
|`DEBUG`          |No |Will set logging level to debug|
|`TEMP_DIR`       |No |Used for downloads and video converting. Bot cleans them up all the time|

### Run in Docker

```shell script
docker run \
  --mount type=volume,target=/tmp \
  -e SLACK_API_TOKEN=xoxb-... \ 
  -e TEMP_DIR=/tmp \
  kozlice/slack-webm-sentinel
```
