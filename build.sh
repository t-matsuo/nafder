#!/bin/bash

OUTPUT_FILE_NAME=nafder
version=$(git describe --tags --abbrev=0)
revision=""
commit=$(git rev-parse --short HEAD)
tag_commit=$(git rev-parse --short $version)

if [ "$commit" != "$tag_commit" ]; then
    revision="-dev"
fi

echo "version: $version"
echo

SCRIPT_DIR=$(cd $(dirname $0); pwd)
[ -d "$SCRIPT_DIR/_build/" ] || mkdir $SCRIPT_DIR/_build

OS_ARCH="darwin_amd64 darwin_arm64 linux_amd64 linux_arm64"

if [ "$GO_OS_ARCH" != "" ]; then
    OS=`echo $GO_OS_ARCH | cut -d "_" -f 1`
    ARCH=`echo $GO_OS_ARCH | cut -d "_" -f 2`
    CGO_ENABLED=0 GOOS=$OS GOARCH=$ARCH go build -ldflags="-X main.version=${version}${revision} -X main.commit=$commit -s -w -buildid=" -trimpath -o $SCRIPT_DIR/_build/${OUTPUT_FILE_NAME}_${GO_OS_ARCH} .
    exit $?
fi

for i in $OS_ARCH; do
    OS=`echo $i | cut -d "_" -f 1`
    ARCH=`echo $i | cut -d "_" -f 2`
    echo $OS $ARCH
    CGO_ENABLED=0 GOOS=$OS GOARCH=$ARCH go build -ldflags="-X main.version=${version}${revision} -X main.commit=$commit -s -w -buildid=" -trimpath -o $SCRIPT_DIR/_build/${OUTPUT_FILE_NAME}_$i .
done
