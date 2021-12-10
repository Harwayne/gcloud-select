# gcloud-select
Simple TUI to select gcloud configurations.

### Development

#### Building

Build with `go build -ldflags="-s -w" gcloud-select.go`. The
[linker flags](https://pkg.go.dev/cmd/link) remove some debug information from
the binary, making it smaller.
