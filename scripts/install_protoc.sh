#!/bin/sh

set -e

PACKAGES="git curl build-base autoconf automake libtool"

apk add --update $PACKAGES

git clone https://github.com/google/protobuf -b $PROTOBUF_TAG --depth 1

cd ./protobuf

./autogen.sh || exit 1
./configure --prefix=/usr || exit 1
make -j 3 || exit 1
make check || exit 1
make install || exit 1

cd ..
rm -rf ./protobuf

apk add --update libstdc++

go get -u -v github.com/golang/protobuf/proto || exit 1
go get -u -v github.com/golang/protobuf/protoc-gen-go || exit 1
go get -u -v google.golang.org/grpc || exit 1

apk del $PACKAGES
rm -rf /var/cache/apk/*