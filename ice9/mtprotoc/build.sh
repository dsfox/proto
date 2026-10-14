#!/bin/bash
# Makes proto/mtproto from proto/ice9/schema (docs/carrying-upstream.md): the
# thirteen files our mtprotoc writes, then the ten .pb.go that protoc and its
# two plugins write from the .proto files, at the versions toolchain.sh pins.
# tests/test_proto_generator_reproduces_the_tree.py builds into a directory of
# its own and requires every file to be the tree's, byte for byte.
#
#   proto/ice9/mtprotoc/build.sh            # into mtproto/
#   proto/ice9/mtprotoc/build.sh <dir>      # into a directory of its own
set -euo pipefail
HERE="$(cd "$(dirname "$0")" && pwd)"
PROTO="$(cd "$HERE/../.." && pwd)"
OUT="${1:-$PROTO/mtproto}"
if ! "$HERE/toolchain.sh" check >/dev/null 2>&1; then
  echo "build.sh: the tools are not at the pinned versions; run $HERE/toolchain.sh" >&2
  "$HERE/toolchain.sh" check >&2 || true
  exit 1
fi
mkdir -p "$OUT"
(cd "$PROTO" && go run ./ice9/mtprotoc generate -schema ice9/schema -out "$OUT")
PB="$(mktemp -d)"
trap 'rm -rf "$PB"' EXIT
(cd "$OUT" && PATH="$HERE/.bin:$PATH" protoc -I=. -I="$HERE/.bin/protoc-29.3/include" \
  --go_out="$PB" --go-grpc_out=require_unimplemented_servers=false:"$PB" schema.tl.*.proto)
cp "$PB"/github.com/teamgram/proto/mtproto/*.pb.go "$OUT/"
echo "build.sh: wrote $(ls "$OUT"/schema.tl.* "$OUT"/codec_schema.tl.pb.go "$OUT"/class_*registers*.go "$OUT"/rpc_client_registers.go | wc -l | tr -d ' ') files into $OUT"
