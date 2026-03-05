# Contributing to slap

Thanks for your interest in contributing! Here's how to get involved.

## Reporting Bugs

Open a [GitHub issue](https://github.com/ajbeck/slack-stdout-pipe/issues) with:

- What you ran (command, OS, Go version)
- What you expected
- What actually happened
- Any relevant logs or error messages

## Suggesting Features

Open an issue describing the use case and why it would be useful. Discussion before code saves everyone time.

## Development Setup

```sh
git clone https://github.com/ajbeck/slack-stdout-pipe.git
cd slack-stdout-pipe
make
```

Requirements:

- Go 1.26+
- GNU Make

## Making Changes

1. Fork the repo and create a branch from `main`.
2. Make your changes.
3. Run `make lint` and `make test` to verify.
4. Open a pull request against `main`.

Keep pull requests focused — one change per PR is easier to review.

## Code Style

- All Go code must pass `go fmt` and `go vet`.
- Follow the conventions already in the codebase.
- Prefer standard library over external dependencies.

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
