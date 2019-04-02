#!/bin/bash

source /root/.pass

docker stop baxx-send-email
docker rm baxx-send-email

docker run \
       --net=host \
       --name=baxx-send-email \
       -e BAXX_SLACK_PANIC="$BAXX_SLACK_PANIC" \
       -e BAXX_POSTGRES="$BAXX_POSTGRES" \
       -e BAXX_SENDGRID_KEY="$BAXX_SENDGRID_KEY" \
       jackdoe/baxx:1.7 /baxx/send_email_queue -debug
