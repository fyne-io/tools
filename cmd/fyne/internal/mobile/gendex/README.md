# gendex

Tool to generate `dex.go` file by compiling Java code residing in the Fyne
source directory (tools directory sibling called `fyne`).

## Generate with Go

Usually run via `go generate ./cmd/fyne/internal/mobile/...`.

## Generate manually

To use a different Fyne source directory or enable verbose mode:

```console
cd ./cmd/fyne/internal/mobile/
go run gendex/gendex.go -o dex.go -s ~/src/fyne -v
cd -
go install ./cmd/fyne
```
