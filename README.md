# slackblr
Slack bot which posts images to a tumblr blog

# Implementation
Quick and dirty

# Config
All configuration is passed in via environment variables

| Variable name        | Description                                  | Example                                            |
| -------------------- | -------------------------------------------- | -------------------------------------------------- |
| PORT                 | Port to listen on                            | 8000                                               |
| SLACK_SIGNING_SECRET | Signing secret for the slack app             | 8f742231b10e8888abcd99yyyzzz85a5                   |
| SLACK_TOKEN          | Slack bot token                              | xoxb-11111111111-aaaaaaaaaaaaaaaaaaaaaaaa          |
| SLACK_CHANNEL        | Channel ID to post images to                 | C0XXXXXXX                                          |
| SLACK_COMMAND        | Slack slash command to respond to            | /tumbly                                            |
| SLACK_USERNAME       | Username to post as                          | tumbly                                             |
| SLACK_USERS          | JSON map of user IDs to nicknames            | {"U0XXXXXXX":"dan","U0ZZZZZZZ":"jon"}              |
| TUMBLR_BLOG          | Tumblr blog name to post to                  | tumblyposts                                        |
| TUMBLR_KEY           | The OAuth consumer key for the Tumblr app    | bGxiAiF8wif1bwErrY7gj2bDxgFaOgkKnV74NxAHWIaWg1L0vz |
| TUMBLR_SECRET        | The OAuth consumer secret for the Tumble app | dWAjG6y3SaODas9AvGUhOkeANhWSgZSHWSCk72x0IBQHh9xMlN |
| TUMBLR_TOKEN         | The Tumblr app token                         | gbyXdlcgMRjcxuOpV9QzqqdTU2438DaEOXqItp33KM9p99Bp3m |
| TUMBLR_TOKEN_SECRET  | The Tumblr app secret                        | 5oHCqNpRH5v1B2P51SoxOlt5Om8n5UXZAntBJpgjHhwPEEI6PO |
