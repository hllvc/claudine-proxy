# Contributing

## Getting Started

Prerequisites: Go 1.25+ with `GOEXPERIMENT=jsonv2` (required for `encoding/json/jsontext` streaming)

```bash
go mod tidy     # Install dependencies
make test       # Run tests
make lint       # Check code style
make bench      # Run benchmarks
make build      # Build binaries
```

## Commit Messages

Use [Conventional Commits](https://www.conventionalcommits.org/).

## Code Standards

- Run `make fmt` and `make lint` before committing
- Add tests for new functionality
- Run `make bench` when modifying streaming/proxy paths
- Prefer stdlib over external dependencies
- Simple > Clever. Explicit > Implicit.

## Important Notes

- `GOEXPERIMENT=jsonv2` is mandatory - we use `jsontext` for streaming JSON transformation
- Benchmark performance-critical paths (SSE streaming, request proxying)

## Questions?

Open a GitHub issue or start a discussion.
