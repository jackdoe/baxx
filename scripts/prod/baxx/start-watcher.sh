#!/bin/bash

source /root/.pass

docker stop baxx-watcher
docker rm baxx-watcher

docker run \
       --net=host \
       --name baxx-watcher \
       -e BAXX_SLACK_PANIC="$BAXX_SLACK_PANIC" \
       -e BAXX_SLACK_MONITORING="$BAXX_SLACK_MONITORING" \
       -e BAXX_POSTGRES="$BAXX_POSTGRES" \
       jackdoe/baxx:2.0.6 /baxx/who_watches_the_watchers
