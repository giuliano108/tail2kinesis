#!/bin/bash

PROJECT=tail2kinesis
WORKDIR=/$PROJECT

cd $(dirname $0)
mkdir $WORKDIR 2>/dev/null
tar -v --exclude /src/.git \
    -c /src | tar --strip-components 1 -C $WORKDIR -xf -
cd $WORKDIR

if [ ! -e .depsfetched ] ; then
    go get -d ./...
    touch .depsfetched
fi
if [ ! -e $GOPATH/src/$PROJECT ] ; then
    ln -s $WORKDIR $GOPATH/src/$PROJECT
fi


make build

echo 'done.'
