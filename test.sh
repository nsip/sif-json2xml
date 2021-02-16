#!/bin/bash
set -e

# Text foreground colour (xterm)
R=`tput setaf 1`
G=`tput setaf 2`
Y=`tput setaf 3`
B=`tput setaf 4`
W=`tput sgr0`

printf "\n"

ip="localhost:1325/"             ###
# XXX Change version
base=$ip"sif-json2xml/convert"    ###

title='SIF-JSON2XML all APIs'
url=$ip
scode=`curl --write-out "%{http_code}" --silent --output /dev/null $url`
if [ $scode -ne 200 ]; then
    echo "${R}${title}${W}"
    exit 1
else
    echo "${G}${title}${W}"
fi
echo "curl $url"
curl -i $url
printf "\n"

## exit 0 ###

sv=3.4.8.draft

SIFJFile=./data/examples/StudentPersonals@3.4.8.draft.json
title='Convert Test @ '$SIFJFile
url=$base"?sv=$sv&wrap" ###
file="@"$SIFJFile
scode=`curl -X POST $url --data-binary $file -w "%{http_code}" -s -o /dev/null`
if [ $scode -ne 200 ]; then
    echo "${R}${title}${W}"
    exit 1
else
    echo "${G}${title}${W}"
fi

xmlname=`basename $SIFJFile .json`.xml
outdir=./data/output/
mkdir -p $outdir
outfile=$outdir"$xmlname"
echo "curl -X POST $url --data-binary $file"
curl -X POST $url --data-binary $file > $outfile
cat $outfile
printf "\n"

echo "${G}All Done${W}"

#######################################################

# ip="localhost:1325/"        ###
# # XXX URL was v0.1.0 - however, the URL returned by config system is 0.0.0 (no v, no 1)
# # XXX Review use of Version 3 digits here, since we need to upgrade last every release, it means
# # all libraries, and all systems always have to upgrade, maybe just v0.1/
# base=$ip"sif-json2xml/v0.1.2"    ###

# title='SIF-JSON2XML all APIs'
# url=$ip
# scode=`curl --write-out "%{http_code}" --silent --output /dev/null $url`
# if [ $scode -ne 200 ]; then
#     echo "${R}Error getting root information from ${ip} - ${title}${W}"
#     exit 1
# else
#     echo "${G}Server OK: ${title}${W}"
# fi
# echo "# Headers: curl $url"
# curl -i $url
# printf "\n"

# ## exit 0 ###

# sv=3.4.7

# JSONFiles=./data/examples/$sv/*
# for f in $JSONFiles
# do
#     title='2XML Test @ '$f
#     url=$base"?sv=$sv" ###
#     file="@"$f
#     scode=`curl -X POST $url --data-binary $file -w "%{http_code}" -s -o /dev/null`
#     if [ $scode -ne 200 ]; then
#         # TODO - be good to get actual text - e.g. 404, not found
#         echo "${R}Error posting binary data ${file} to ${url} - ${title} code=${scode}${W}"
#         exit 1
#     else
#         echo "${G}${title}${W}"
#     fi

#     sifname=`basename $f .json`.xml
#     outdir=./data/output/$sv/
#     mkdir -p $outdir
#     outfile=$outdir"$sifname"
#     echo "curl -X POST $url --data-binary $file"
#     curl -X POST $url --data-binary $file > $outfile
#     cat $outfile
#     printf "\n"
# done

# echo "${G}All Done${W}"
