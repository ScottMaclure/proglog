# ProgLog

## Setup

```bash
# From project-root

go get github.com/gogo/protobuf/...@v1.3.1

protoc api/v1/*.proto \
--gogo_out=Mgogoproto/gogo.proto=github.com/gogo/protobuf/proto:. \
--proto_path=$(go list -f '{{ .Dir }}' -m github.com/gogo/protobuf) \
--proto_path=.

```