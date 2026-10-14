#!/bin/bash
# The tools that made proto/mtproto's generated files, at the versions their
# headers name: protoc v5.29.3 and protoc-gen-go v1.36.5 (schema.tl.sync.pb.go),
# protoc-gen-go-grpc v1.5.1 (schema.tl.sync_service_grpc.pb.go), and clang-format
# for the .proto files (build2.sh). Any other version writes another header and
# may lay the descriptors out differently, and the test that compares our output
# with the tree byte for byte would then measure the tools, not the generator.
# Everything lands in .bin/ beside this script, outside git.
#
#   proto/ice9/mtprotoc/toolchain.sh          # install what is missing, then check
#   proto/ice9/mtprotoc/toolchain.sh check    # only say whether every version is right
set -euo pipefail
cd "$(dirname "$0")"
BIN="$PWD/.bin"
PROTOC_VERSION=29.3
GEN_GO_VERSION=v1.36.5
GEN_GRPC_VERSION=v1.5.1
PROTOC_ZIP="protoc-$PROTOC_VERSION-osx-aarch_64.zip"
PROTOC_URL="https://github.com/protocolbuffers/protobuf/releases/download/v$PROTOC_VERSION/$PROTOC_ZIP"

install() {
  mkdir -p "$BIN"
  if [ ! -x "$BIN/protoc-$PROTOC_VERSION/bin/protoc" ]; then
    echo "▸ protoc $PROTOC_VERSION"
    curl -sSL -o "$BIN/$PROTOC_ZIP" "$PROTOC_URL"
    rm -rf "$BIN/protoc-$PROTOC_VERSION"
    unzip -q "$BIN/$PROTOC_ZIP" -d "$BIN/protoc-$PROTOC_VERSION"
    rm -f "$BIN/$PROTOC_ZIP"
  fi
  ln -sf "protoc-$PROTOC_VERSION/bin/protoc" "$BIN/protoc"
  if [ ! -x "$BIN/protoc-gen-go" ] || ! "$BIN/protoc-gen-go" --version 2>/dev/null | grep -q " $GEN_GO_VERSION\$"; then
    echo "▸ protoc-gen-go $GEN_GO_VERSION"
    GOBIN="$BIN" go install "google.golang.org/protobuf/cmd/protoc-gen-go@$GEN_GO_VERSION"
  fi
  if [ ! -x "$BIN/protoc-gen-go-grpc" ] || ! "$BIN/protoc-gen-go-grpc" --version 2>/dev/null | grep -q " ${GEN_GRPC_VERSION#v}\$"; then
    echo "▸ protoc-gen-go-grpc $GEN_GRPC_VERSION"
    GOBIN="$BIN" go install "google.golang.org/grpc/cmd/protoc-gen-go-grpc@$GEN_GRPC_VERSION"
  fi
  if ! command -v clang-format >/dev/null; then
    echo "▸ clang-format (brew)"
    brew install clang-format >/dev/null
  fi
}

check() {
  local failed=0
  expect() {
    local name="$1" want="$2" got="$3"
    if [ "$got" = "$want" ]; then
      echo "  ok   $name: $got"
    else
      echo "  FAIL $name: $got, want $want" >&2
      failed=1
    fi
  }
  expect protoc "libprotoc $PROTOC_VERSION" "$("$BIN/protoc" --version 2>/dev/null || echo missing)"
  expect protoc-gen-go "protoc-gen-go $GEN_GO_VERSION" "$("$BIN/protoc-gen-go" --version 2>/dev/null || echo missing)"
  expect protoc-gen-go-grpc "protoc-gen-go-grpc ${GEN_GRPC_VERSION#v}" "$("$BIN/protoc-gen-go-grpc" --version 2>/dev/null || echo missing)"
  if command -v clang-format >/dev/null; then
    echo "  ok   clang-format: $(clang-format --version | head -1) (teamgram's version is not recorded; the byte test decides)"
  else
    echo "  FAIL clang-format: missing" >&2
    failed=1
  fi
  return $failed
}

case "${1:-install}" in
  install) install; check ;;
  check) check ;;
  *) echo "usage: toolchain.sh [install|check]" >&2; exit 2 ;;
esac
