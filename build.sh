#!/bin/bash
# rm -f ./go.sum
# go get -u ./...

R=`tput setaf 1`
G=`tput setaf 2`
Y=`tput setaf 3`
W=`tput sgr0`

ORIPATH=`pwd`

cd ./config && ./build.sh && cd "$ORIPATH" 
echo "${G}CONFIG PREPARED${W}"

cd ./server && ./build.sh && cd "$ORIPATH" 
echo "${G}SERVER BUILDING DONE${W}"
