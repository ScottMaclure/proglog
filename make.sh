#!/usr/bin/bash
set -eu

echo "Compiling proto3 files..."

# Option 1: gogoproto proto3 compile
# protoc api/v1/*.proto \
#     --gogo_out=Mgogoproto/gogo.proto=github.com/gogo/protobuf/proto:. \
#     --proto_path=$(go list -f '{{ .Dir }}' -m github.com/gogo/protobuf) \
#     --proto_path=.

# Option 2: Use the default compiler, which got an update recently.
# https://developers.google.com/protocol-buffers/docs/gotutorial
protoc -I=api/v1 --go_out=$GOPATH/src ./api/v1/*.proto

echo "Done"
