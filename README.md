# Slack webm sentinel

Slack doesn't have an inline player for .webm links.

This bot fixes the problem: it tracks .webm links in messages, downloads them, converts into .mp4 and posts them back
into the channel.

### Configuration

Done via env variables:

|Name|Required?|Description|
|---|---|---|
|`SLACK_API_TOKEN`|Yes|Bot access token (starts with `xoxb-...`)|
|`NOTIFY_MODE`    |No |How to let users know about bot activity. Can be one of:<ul><li>`reaction` - default, will add emoji to a message with link)</li><li>`message` - will post a message when it finds URL</li><li>`none` - do not notify</li>|
|`DEBUG`          |No |Will set logging level to debug|
|`TEMP_DIR`       |No |Will be used for downloads and video converting. Bot removes them after link is handled. Default is `/tmp`|

### Run in Docker

```shell script
docker run \
  --mount type=volume,target=/tmp \
  -e SLACK_API_TOKEN=xoxb-... \ 
  kozlice/slack-webm-sentinel
```
