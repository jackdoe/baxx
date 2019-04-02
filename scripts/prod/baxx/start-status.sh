#!/bin/bash
source /root/.pass
docker stop baxx-status
docker rm baxx-status
docker run \
       --net=host \
       --privileged \
       --name baxx-status \
       -e BAXX_SLACK_PANIC="$BAXX_SLACK_PANIC" \
       -e BAXX_SLACK_MONITORING="$BAXX_SLACK_MONITORING" \
       -e BAXX_POSTGRES="$BAXX_POSTGRES" \
       jackdoe/baxx:1.7 /baxx/status -debug
