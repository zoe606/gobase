#!/usr/bin/env bash
set -euo pipefail

case "$(uname -s)-$(uname -m)" in
  Linux-x86_64) platform=linux-x86_64; checksum=121f6c7afe1d4d0e3ea6aab9432038599250134cbf4474cb1167d2c7decd4278 ;;
  Linux-aarch64|Linux-arm64) platform=linux-aarch_64; checksum=8b8f18bd2b30346efbc698dd5a73dd7c805f3ef8380f6dfc95c768f3f1852f6a ;;
  Darwin-arm64) platform=osx-aarch_64; checksum=9cd98a532c5c5e0c4161314de0225de27e4c8a323917b6ea7b1b714d3ae23466 ;;
  Darwin-x86_64) platform=osx-x86_64; checksum=228cc7add4616cc14ca5e80dee83209d44449a7aee95a914ae748fa374efb078 ;;
  *) echo 'Protobuf generation supports Linux/macOS on amd64/arm64' >&2; exit 1 ;;
esac

directory="bin/protoc-36.2/$platform"
if [ ! -x "$directory/bin/protoc" ]; then
  mkdir -p "$directory"
  archive=$(mktemp)
  trap 'rm -f "$archive"' EXIT
  curl -fsSL "https://github.com/protocolbuffers/protobuf/releases/download/v36.2/protoc-36.2-$platform.zip" -o "$archive"
  if command -v sha256sum >/dev/null; then
    actual=$(sha256sum "$archive" | cut -d ' ' -f 1)
  else
    actual=$(shasum -a 256 "$archive" | cut -d ' ' -f 1)
  fi
  test "$actual" = "$checksum"
  unzip -q -o "$archive" -d "$directory"
fi

module=$(go list -m)
gen_go=$(go tool -n protoc-gen-go)
gen_grpc=$(go tool -n protoc-gen-go-grpc)
proto=pkg/api/echo/v1/echo.proto
mapping="$proto=$module/pkg/api/echo/v1;echov1"
"$directory/bin/protoc" --plugin="protoc-gen-go=$gen_go" --plugin="protoc-gen-go-grpc=$gen_grpc" \
  --go_out=. --go_opt=paths=source_relative --go_opt="M$mapping" \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative --go-grpc_opt="M$mapping" "$proto"
