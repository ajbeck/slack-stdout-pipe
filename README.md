# slap

`slap` (Slack Application Pipe) wraps any terminal command and streams its stdout and stderr to a Slack channel via an incoming webhook.

Output appears in your terminal in real time and is batched into Slack messages with `[stdout]` and `[stderr]` tags so you can tell them apart. A start message, streamed output, and a final summary with exit code and duration are posted automatically.

## Prerequisites

- Go 1.26+
- A Slack workspace with an incoming webhook URL

## Slack Webhook Setup

1. Go to [api.slack.com/apps](https://api.slack.com/apps) and create a new app (or use an existing one).
2. Enable **Incoming Webhooks** under your app's settings.
3. Click **Add New Webhook to Workspace** and select the channel you want output sent to.
4. Copy the webhook URL — it looks like `https://hooks.slack.com/services/T.../B.../xxx`.

For full details see the [Slack incoming webhooks documentation](https://docs.slack.dev/messaging/sending-messages-using-incoming-webhooks/).

## Install

```sh
go install github.com/ajbeck/slack-stdout-pipe/cmd/slap@latest
```

Or build from source:

```sh
git clone https://github.com/ajbeck/slack-stdout-pipe.git
cd slack-stdout-pipe
make slap
# binary is at ./bin/slap
```

## Usage

Set the `SLAP_TARGET` environment variable to your Slack webhook URL, then prefix any command with `slap`:

```sh
export SLAP_TARGET="https://hooks.slack.com/services/T.../B.../xxx"

slap ls -la /tmp
slap make test
slap ./my-script.sh --verbose
```

The wrapped command's exit code is preserved — `slap` exits with whatever the child process returned.

### What appears in Slack

**Start message:**
> :rocket: \`ls -la /tmp\` started

**Streamed output** (batched every 500ms):
```
[stdout] drwxr-xr-x  5 user  staff  160 Mar  5 10:00 .
[stdout] -rw-r--r--  1 user  staff   42 Mar  5 09:59 foo.txt
[stderr] ls: /tmp/bar: Permission denied
```

**Final message:**
> :white_check_mark: \`ls -la /tmp\` exited 0 in 1.23s

Or on failure:
> :x: \`ls -la /tmp\` exited 1 in 0.45s

## Demo

A sample app is included that outputs lines from *The Adventures of Sherlock Holmes* to stdout and stderr at random intervals. It's useful for testing `slap` without a real workload.

Build it:

```sh
make demo
```

Run it through `slap` for 30 seconds:

```sh
export SLAP_TARGET="https://hooks.slack.com/services/T.../B.../xxx"
slap ./bin/demo 30s
```

The demo accepts a single argument — the duration to run (e.g. `10s`, `1m`, `2m30s`).

## Environment Variables

| Variable | Required | Description |
|---|---|---|
| `SLAP_TARGET` | Yes | Slack incoming webhook URL. Removed from the child process environment so wrapped commands don't see it. |

## Build Targets

```sh
make          # build slap and demo into ./bin/
make slap     # build just slap
make demo     # build just demo
make vet      # run go vet
make test     # run go test
make clean    # remove ./bin/ and clear go build cache
```
