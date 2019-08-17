#!/bin/bash

PROJECT=tail2kinesis
WORKDIR=/$PROJECT

cd $(dirname $0)
mkdir $WORKDIR 2>/dev/null
tar -v --exclude /src/.git \
    -c /src | tar --strip-components 1 -C $WORKDIR -xf -
cd $WORKDIR

make build

echo 'done.'
