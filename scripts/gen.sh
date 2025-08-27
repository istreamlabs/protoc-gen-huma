#!/bin/bash -e

SCRIPT_ROOT="$(cd "$(dirname "$0")"; pwd)"

# Ensure the *.proto globs expand in the same order, overriding the user's locale.
export LC_ALL=C

CMD="run_docker"
if [[ "$1" == "codefresh" ]];
then
  echo "running protoc directly..."
  unset CMD
fi

run_docker() {
  DOCKER_USER=""
  if [[ "$OSTYPE" == "linux-gnu" ]]; then
    DOCKER_USER="-u `id -u`:`id -g`"
  fi

  # See https://github.com/iStreamPlanet/docker-protoc for details on
  # istreamplanet/protoc.
  docker run --rm -v `pwd`:/host $DOCKER_USER istreamplanet/protoc:1.9.0 $*
}

cd "$SCRIPT_ROOT"

# Remove previously generated files.
rm -rf go tmp 2> /dev/null || true
find ./doc/*.dot* -delete || true

GO_OUT_DIR=${SCRIPT_ROOT}/go
PROTOC_INCLUDE="-I ./example/proto -I /protobuf/include"

echo "Generating Go protobuf"
mkdir tmp
PROTOC_GO="$PROTOC_INCLUDE --go_out=plugins=grpc:tmp --twirp_out=tmp --validate_out=lang=go:tmp"
$CMD protoc $PROTOC_GO example/proto/package2/*.proto
$CMD protoc $PROTOC_GO example/proto/package1/*.proto

rm -rf example/go
mv tmp/github.com/istreamlabs/protoc-gen-huma/example/go example
rm -rf tmp

echo "Generating Go Huma bindings"
$CMD protoc --huma_out=example/go \
  $PROTOC_INCLUDE \
  ./example/proto/package1/*.proto \
  ./example/proto/package2/*.proto
