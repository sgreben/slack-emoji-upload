# slack-emoji-upload

Upload custom Slack emoji from the CLI.

<!-- TOC -->

- [Get it](#get-it)
- [Use it](#use-it)
- [Authentication](#authentication)
  - [Token](#token)
  - [Password](#password)
- [Example](#example)
  - [Token auth](#token-auth)
  - [Password auth](#password-auth)

<!-- /TOC -->


## Get it

- Either download the statically linked binary from [the latest release](https://github.com/sgreben/slack-emoji-upload/releases/latest)

- ...or use `go get`:
    ```sh
    go get github.com/sgreben/slack-emoji-upload
    ```

## Use it

```text
slack-emoji-upload OPTIONS [FILES]

Options:
  -token string
        Slack API token
  -email string
        user email (required when -token not specified)
  -password string
        user password (required when -token not specified)
  -team string
        Slack team (required when -token not specified)
  -rate-limit duration
        upload rate limit (1 emoji per ...) (default 2s)

  -notify-channel string
        notify this channel on successful uploads
  -quiet
        suppress log output
```

## Authentication

### Token

To authenticate with a token (`-token` option), you need to use a `xoxs-*` Slack API token, not a regular user token. It looks something like this:

```
xoxs-abcdefghij-klmnopqrstuv-wxyzabcdefgh-ijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrst
```

[**How to obtain the Slack API token**](https://github.com/jackellenberger/emojme#finding-a-slack-token)

### Password

Alternatively, you can provide your `-team`, `-email` and `-password` to let the app obtain a `xoxs-*` API token for you.

## Example

### Token auth

```sh
$ slack-emoji-upload -token "$MY_TOKEN" emoji/*.*
2018/09/11 11:34:53 reeeeee: uploading "/tmp/emoji/reeeeee.gif"
2018/09/11 11:34:55 reeeeee: uploaded
2018/09/11 11:34:55 yeet: uploading "/tmp/emoji/yeet.png"
2018/09/11 11:34:57 yeet: uploaded
```

### Password auth

```sh
$ slack-emoji-upload -team my-team -email "me@example.com" -password "hunter2" emoji/*.*
2018/09/11 11:34:53 reeeeee: uploading "/tmp/emoji/reeeeee.gif"
2018/09/11 11:34:55 reeeeee: uploaded
2018/09/11 11:34:55 yeet: uploading "/tmp/emoji/yeet.png"
2018/09/11 11:34:57 yeet: uploaded
```
