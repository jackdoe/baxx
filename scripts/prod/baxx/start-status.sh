#!/bin/bash

docker stop baxx-status
docker stop
docker run \
       --net=host \
       --privileged \
       --name baxx-status \
       -e BAXX_SLACK_PANIC="$BAXX_SLACK_PANIC" \
       -e BAXX_SLACK_MONITORING="$BAXX_SLACK_MONITORING" \
       -e BAXX_POSTGRES="$BAXX_POSTGRES" \
       jackdoe/baxx:1.2 /baxx/notification_run -debug
