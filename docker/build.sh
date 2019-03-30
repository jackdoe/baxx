#!/bin/bash
SCRIPT=$(readlink -f "$0")
dir=$(dirname "$SCRIPT")

mkdir -p $dir/bin
echo rm -rf $dir/t && cp -rp $dir/../help/t $dir/t

cd $dir/../cmd/notification_run/
go build -o $dir/bin/notification_run
cd -

cd $dir/../cmd/send_email_queue/
go build -o $dir/bin/send_email_queue
cd -

cd $dir/../api/
go build -o $dir/bin/api
cd -

