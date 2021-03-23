#!/bin/bash
# rm -f ./go.sum
# go get -u ./...

ORIPATH=`pwd`

cd ./config && ./build_d.sh && cd "$ORIPATH" 
echo "CONFIG PREPARED"

cd ./server && ./build_d.sh && cd "$ORIPATH" 
echo "SERVER BUILDING DONE"
