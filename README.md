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

Download a prebuilt binary from the [latest release](https://github.com/ajbeck/slack-stdout-pipe/releases/latest):

| Platform | Asset |
|---|---|
| macOS Apple Silicon | `slap-darwin-arm64.tar.gz` |
| macOS Intel | `slap-darwin-amd64.tar.gz` |
| Linux x64 | `slap-linux-amd64.tar.gz` |
| Linux ARM64 | `slap-linux-arm64.tar.gz` |

```sh
tar -xzf slap-<os>-<arch>.tar.gz
mv slap-<os>-<arch> /usr/local/bin/slap
```

Or install with `go install`:

```sh
go install github.com/ajbeck/slack-stdout-pipe/cmd/slap@latest
```

Or build from source:

```sh
git clone https://github.com/ajbeck/slack-stdout-pipe.git
cd slack-stdout-pipe
make build-slap
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
> :rocket: \`ls -la /tmp\` started — slap 0.1.0+local:2025-03-05T12:00:00Z

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
make build-demo
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
make              # build slap and demo into ./bin/
make build-slap   # build just slap
make build-demo   # build just demo
make test         # run tests
make vet          # run go vet
make lint         # run fmt and vet
make clean        # remove ./bin/ and clear test cache
```
