# ProgLog

## Setup

```bash
# From project-root

# Put into my ~/.bashrc
export GOROOT=/c/Go
export GOPATH=/E/MEGA/dev/go
export GOBIN=$GOPATH/bin
export PATH=$PATH:$GOROOT:$GOPATH:$GOBIN

go get github.com/gogo/protobuf/...@v1.3.1

protoc api/v1/*.proto \
--gogo_out=Mgogoproto/gogo.proto=github.com/gogo/protobuf/proto:. \
--proto_path=$(go list -f '{{ .Dir }}' -m github.com/gogo/protobuf) \
--proto_path=.

# Using Google's compiler

go install google.golang.org/protobuf/cmd/protoc-gen-go

protoc -I=api/v1 --go_out=api/v1 *.proto

```