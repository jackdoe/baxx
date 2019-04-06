#!/bin/bash

source /root/.pass

docker stop baxx-notification-rules
docker rm baxx-notification-rules

docker run \
       --net=host \
       --name=baxx-notification-rules \
       -e BAXX_SLACK_PANIC="$BAXX_SLACK_PANIC" \
       -e BAXX_POSTGRES="$BAXX_POSTGRES" \
       jackdoe/baxx:2.0.0 /baxx/notification_run