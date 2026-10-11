#!/bin/bash
# Puts ice9's lines back into the two files the schema generator writes (#198):
# mtproto/class_name_registers.go (a client ahead of the schema is served what
# we have) and mtproto/codec_schema.tl.pb.go (updateMlsMailbox, hand-written in
# mls_update.go). Upstream regenerates both thirty-odd times a year, so at a
# carry take upstream's two files and run this. tests/test_proto_codegen_patch.py
# checks that the base plus this patch is exactly what the tree holds.
set -euo pipefail
cd "$(dirname "$0")/.."
if git apply --reverse --check ice9/codegen.patch 2>/dev/null; then
  echo "ice9/codegen.patch is already applied"
  exit 0
fi
git apply ice9/codegen.patch
echo "ice9/codegen.patch applied"
