# AILERON Observability

**This repository provides observability features**.

<div align="center">

[![GoDoc](https://godoc.org/github.com/aileron-projects/aileron-observability?status.svg)](http://godoc.org/github.com/aileron-projects/aileron-observability)
[![Go Report Card](https://goreportcard.com/badge/github.com/aileron-projects/aileron-observability)](https://goreportcard.com/report/github.com/aileron-projects/aileron-observability)
[![License](https://img.shields.io/badge/License-Apache%202.0-yellow.svg)](./LICENSE)

[![Test Suite](https://github.com/aileron-projects/aileron-observability/actions/workflows/test-suite.yaml/badge.svg?branch=main)](https://github.com/aileron-projects/aileron-observability/actions/workflows/test-suite.yaml?query=branch%3Amain)
[![Check Suite](https://github.com/aileron-projects/aileron-observability/actions/workflows/check-suite.yaml/badge.svg?branch=main)](https://github.com/aileron-projects/aileron-observability/actions/workflows/check-suite.yaml?query=branch%3Amain)
[![OpenSourceInsight](https://badgen.net/badge/open%2Fsource%2F/insight/cyan)](https://deps.dev/aileron-observability/github.com%2Faileron-projects%2Faileron-observability)
[![OSS Insight](https://badgen.net/badge/OSS/Insight/orange)](https://ossinsight.io/analyze/aileron-projects/aileron-observability)

</div>

AI generated docs are available at:

- [DeepWiki](https://deepwiki.com/aileron-projects/aileron-observability)
- [GitDiagram](https://gitdiagram.com/aileron-projects/aileron-observability)

## Usage

This project is provided as a Go module.

Use go command to use from your project.

```bash
go get github.com/aileron-projects/aileron-observability@latest
go mod tidy
```

## Tested Environment

Operating System:

- `Linux`: [ubuntu-latest](https://github.com/actions/runner-images)
- `Windows`: [windows-latest](https://github.com/actions/runner-images)
- `macOS`: [macos-latest](https://github.com/actions/runner-images)

Go:

- Current Stable: `go 1.(N).x`
- Previous Stable: `go 1.(N-1).x`
- Minimum Requirement: `go 1.(N-2).0`
  - Declared in the [go.mod](go.mod)

Where `N` is the current latest minor version.
See the Go official release page [Stable versions](https://go.dev/dl/).

In addition to the environment above, following platforms are tested on ubuntu
using [QEMU User space emulator](https://www.qemu.org/docs/master/user/main.html).

- `amd64`
- `arm/v5`
- `arm/v6`
- `arm/v7`
- `arm64`
- `ppc64`
- `ppc64le`
- `riscv64`
- `s390x`
- `loong64`
- `386`
- `mips`
- `mips64`
- `mips64le`
- `mipsle`

## Release Cycle

- Releases are made as needed.
- Versions follow [Semantic Versioning](https://semver.org/).
  - `vX.Y.Z`
  - `vX.Y.Z-rc.N`
  - `vX.Y.Z-beta.N`
  - `vX.Y.Z-alpha.N`

## License

Apache 2.0
