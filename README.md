# ProgLog

## Setup

```bash
# From project-root

# Put into my ~/.bashrc
export GOROOT=/c/Go
export GOPATH=/E/MEGA/dev/go
export GOBIN=$GOPATH/bin
export PATH=$PATH:$GOROOT:$GOPATH:$GOBIN

# 1) Using gogoproto (broken for me)

go get github.com/gogo/protobuf/...@v1.3.1

protoc api/v1/*.proto \
--gogo_out=Mgogoproto/gogo.proto=github.com/gogo/protobuf/proto:. \
--proto_path=$(go list -f '{{ .Dir }}' -m github.com/gogo/protobuf) \
--proto_path=.

# 2) Using Go protocol buffers plugin

go install google.golang.org/protobuf/cmd/protoc-gen-go

# From project root
# protoc -I=api/v1 --go_out=api/v1 *.proto 
protoc --go_out=. api/v1/*.proto 

# Ubuntu/Bash
sudo apt install golang-go
export GOROOT=/usr/lib/go
export GOPATH=/mnt/e/MEGA/dev/go
export GOBIN=$GOPATH/bin
export PATH=$PATH:$GOROOT:$GOPATH:$GOBIN
export GO111MODULE=on
sudo apt install protobuf-compiler
go get github.com/gogo/protobuf/...@v1.3.1
go install google.golang.org/protobuf/cmd/protoc-gen-go

```