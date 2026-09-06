#!/usr/bin/env bash
#
# Regenerate the protos that still live in this repo.
#
# Only protos/events does. Both agent contracts moved to modules of their own,
# each with its own gen.sh:
#
#   frppc <-> master   cnb.cool/sixkun/sixfrp-client-proto   (published)
#   frpps <-> master   cnb.cool/sixkun/sixfrp-server-proto   (private)
#
# Both ship generated Go, so nothing here regenerates them. The
# protoc-go-inject-tag pass that used to run over protos/pb went with them — it
# had been a no-op for a while, since no .proto in this repo carries a
# `// @gotags:` comment.
#
# The www TypeScript output is gone too: it was generated from the frpps protos
# into www/protos/pb and never imported by anything under www/src.

set -eu

export GOEXPERIMENT=jsonv2

cd "$(dirname "$0")"

PROTOC_PATH=$(command -v protoc)

# --proto_path=. from the repo root, so the registered file path stays
# protos/events/v1/event.proto. That prefix is what keeps this file from
# colliding with some other module's v1/event.proto in the protobuf registry —
# the same reason the two agent contracts sit behind sixfrp/<side>/.
$PROTOC_PATH --proto_path=. \
  --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  protos/events/v1/*.proto
