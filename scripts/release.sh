#!/bin/sh
set -eu
version=${1:-dev}
case "$version" in *[!a-zA-Z0-9._-]*|'') printf '%s\n' 'Version may contain letters, digits, dots, underscores and hyphens.' >&2; exit 1;; esac
mkdir -p dist
for target in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64; do
  build_os=${target%/*}
  build_arch=${target#*/}
  extension=""
  if [ "$build_os" = windows ]; then extension=.exe; fi
  output="dist/horizon-winner-${version}-${build_os}-${build_arch}${extension}"
  CGO_ENABLED=0 GOOS="$build_os" GOARCH="$build_arch" go build -trimpath -buildvcs=false -ldflags="-s -w -buildid= -X main.version=$version" -o "$output" ./cmd/horizon-winner
  server_output="../dist/singularity-server-${version}-${build_os}-${build_arch}${extension}"
  (cd server && CGO_ENABLED=0 GOOS="$build_os" GOARCH="$build_arch" go build -trimpath -buildvcs=false -ldflags="-s -w -buildid=" -o "$server_output" ./cmd/singularity-server)
done
# Works on macOS and Linux; paths and ordering are fixed.
(cd dist && for build_file in horizon-winner-"$version"-* singularity-server-"$version"-*; do shasum -a 256 "$build_file"; done) > "dist/SHA256SUMS-$version.txt"
