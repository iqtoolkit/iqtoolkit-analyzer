---
sidebar_position: 2
---

# Installation

## Prerequisites

- Go 1.22 or later
- Access to a PostgreSQL instance (for configuration analysis)
- PostgreSQL log files (for log-based analysis)

## Pre-built Binaries

Pre-built binaries (Linux, macOS, Windows — amd64/arm64) are available on the [releases page](https://github.com/iqtoolkit/iqtoolkit-analyzer/releases):

```bash
# macOS arm64
curl -sL https://github.com/iqtoolkit/iqtoolkit-analyzer/releases/latest/download/iqtoolkit-analyzer_Darwin_arm64.tar.gz | tar xz
sudo mv iqtoolkit-analyzer /usr/local/bin/

# macOS amd64
curl -sL https://github.com/iqtoolkit/iqtoolkit-analyzer/releases/latest/download/iqtoolkit-analyzer_Darwin_amd64.tar.gz | tar xz
sudo mv iqtoolkit-analyzer /usr/local/bin/

# Linux amd64
curl -sL https://github.com/iqtoolkit/iqtoolkit-analyzer/releases/latest/download/iqtoolkit-analyzer_Linux_amd64.tar.gz | tar xz
sudo mv iqtoolkit-analyzer /usr/local/bin/
```

## With Go

```bash
go install github.com/iqtoolkit/iqtoolkit-analyzer@latest
```

## From Source

```bash
git clone https://github.com/iqtoolkit/iqtoolkit-analyzer.git
cd iqtoolkit-analyzer
make build
```
