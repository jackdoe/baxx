#!/bin/bash

source /root/.pass

docker stop baxx-api
docker rm baxx-api

docker run \
       --net=host \
       --name=baxx-api \
       -e BAXX_SLACK_PANIC="$BAXX_SLACK_PANIC" \
       -e BAXX_POSTGRES="$BAXX_POSTGRES" \
       -e BAXX_JUDOC_URL="http://localhost:9122/" \
       jackdoe/baxx:1.9.5 /baxx/api -debug
