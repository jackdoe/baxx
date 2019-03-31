source /root/.pass
# this is initial version, move to systemd instead of docker --restart
# so we can set dependencies

VERSION=1.0

#######
# app #
#######
docker run \
       -dit \
       --net=host \
       --restart always \
       --name baxx-api \
       -e BAXX_SLACK_PANIC="$BAXX_SLACK_PANIC" \
       -e BAXX_POSTGRES="$BAXX_POSTGRES" \
       -e BAXX_JUDOC_URL="http://localhost:9122/" \
       jackdoe/baxx:$VERSION /baxx/api -debug

docker run \
       -dit \
       --net=host \
       --restart always \
       --name baxx-send-email \
       -e BAXX_SLACK_PANIC="$BAXX_SLACK_PANIC" \
       -e BAXX_POSTGRES="$BAXX_POSTGRES" \
       -e BAXX_SENDGRID_KEY="$BAXX_SENDGRID_KEY" \
       jackdoe/baxx:$VERSION /baxx/send_email_queue -debug

if [ $(hostname) = "bb.baxx.dev" ]; then
    docker run \
           -dit \
           --net=host \
           --restart always \
           --name baxx-notification-rules \
           -e BAXX_SLACK_PANIC="$BAXX_SLACK_PANIC" \
           -e BAXX_POSTGRES="$BAXX_POSTGRES" \
           jackdoe/baxx:$VERSION /baxx/notification_run -debug
fi

#######################
# status and watchers #
#######################
docker run \
       -dit \
       --net=host \
       --restart always \
       --privileged \
       --name baxx-watcher \
       -e BAXX_SLACK_PANIC="$BAXX_SLACK_PANIC" \
       -e BAXX_SLACK_MONITORING="$BAXX_SLACK_MONITORING" \
       -e BAXX_POSTGRES="$BAXX_POSTGRES" \
       jackdoe/baxx:$VERSION /baxx/who_watches_the_watchers -debug

# needs privileged for mdadm
docker run \
       -dit \
       --net=host \
       --restart always \
       --privileged \
       --name baxx-status \
       -e BAXX_SLACK_PANIC="$BAXX_SLACK_PANIC" \
       -e BAXX_SLACK_MONITORING="$BAXX_SLACK_MONITORING" \
       -e BAXX_POSTGRES="$BAXX_POSTGRES" \
       jackdoe/baxx:$VERSION /baxx/notification_run -debug
