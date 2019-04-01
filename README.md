# baxx.dev

check it out `ssh register@ui.baxx.dev`

[ work in progress ]

# backup service
(also i am learning how to build a product without a website haha)

# screenshots
```
┌────────────────────────────────────────────┐
│                                            │
│ ██████╗  █████╗ ██╗  ██╗██╗  ██╗           │
│ ██╔══██╗██╔══██╗╚██╗██╔╝╚██╗██╔╝           │
│ ██████╔╝███████║ ╚███╔╝  ╚███╔╝            │
│ ██╔══██╗██╔══██║ ██╔██╗  ██╔██╗            │
│ ██████╔╝██║  ██║██╔╝ ██╗██╔╝ ██╗           │
│ ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝           │
│                                            │
│ Storage 10G                                │
│   Trial 1 Month 0.1 EUR                    │
│   Subscription: 5 EUR per Month            │
│   Availability: ALPHA                      │
│                                            │
│ Email                                      │
│ █                                          │
│                                            │
│ Password                                   │
│                                            │
│                                            │
│ Confirm Password                           │
│                                            │
│                                            │
│ Registering means you agree with           │
│ the terms of service!                      │
│                                            │
│                 [Register]                 │
│                                            │
│ [Help]  [What/Why/How]  [Terms Of Service] │
│                                            │
│                   [Quit]                   │
└────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────────────┐
│                                                                          │
│ ██████╗  █████╗ ██╗  ██╗██╗  ██╗                                         │
│ ██╔══██╗██╔══██╗╚██╗██╔╝╚██╗██╔╝                                         │
│ ██████╔╝███████║ ╚███╔╝  ╚███╔╝                                          │
│ ██╔══██╗██╔══██║ ██╔██╗  ██╔██╗                                          │
│ ██████╔╝██║  ██║██╔╝ ██╗██╔╝ ██╗                                         │
│ ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝                                         │
│                                                                          │
│                                                                          │
│ Email: example@example.com                                               │
│ Verification pending.                                                    │
│ Please check your spam folder.                                           │
│                                                                          │
│ Subscription:                                                            │
│ Activate at https://baxx.dev/sub/XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX    │
│                                                                          │
│ Refreshing.. -                                                           │
│                                                                          │
│                [█Help] [Resend Verification Email]  [Quit]               │
└──────────────────────────────────────────────────────────────────────────┘


Hi,

The service I offer is still in Alpha stage, but I really appreciate
the support.

# Subscription

## Plan (only one for now):

Storage 10G
Trial 1 Month 0.1E
Subscription: 5E per Month
Availability: ALPHA

Here be Dragons! Data can be lost!

## Subscribe

In order to use baxx.dev you need a subscription,
At the moment I support only paypal.com, please visit:

https://baxx.dev/sub/WILL-BE-IN-YOUR-EMAIL
To be redirected to paypal.com.

## Verify your email

Email verification is also required, you should've received the
verification link in another email.


Thanks again!

████████╗ ██████╗ ██╗  ██╗███████╗███╗   ██╗███████╗
╚══██╔══╝██╔═══██╗██║ ██╔╝██╔════╝████╗  ██║██╔════╝
   ██║   ██║   ██║█████╔╝ █████╗  ██╔██╗ ██║███████╗
   ██║   ██║   ██║██╔═██╗ ██╔══╝  ██║╚██╗██║╚════██║
   ██║   ╚██████╔╝██║  ██╗███████╗██║ ╚████║███████║
   ╚═╝    ╚═════╝ ╚═╝  ╚═╝╚══════╝╚═╝  ╚═══╝╚══════╝


Tokens are like backup namespaces, you can have the same file in
different tokens and it won't conflict.

There are 2 kinds of tokens, ReadWrite and WriteOnly, ReadWrite tokens
dont require any credentials for create, delete and list files,
WriteOnly tokens require credentials for *list* and *delete*.


## Current Tokens

TOKEN: TOKEN-UUID-A
  Name: db-example-a
  Write Only: false
  Keep N Versions: 3 (per file)
  Alert:
     Name: more than 1 day old database backup
     Matching on Regexp: "\.sql"
     UUID: NOTIFICATION-UUID
     Notify If:
        - size delta is irrelevant
        + last version is older than 1 days

  Alert:
     Name: database file is 50% different
     Matching on Regexp: "\.sql"
     UUID: NOTIFICATION-UUID
     Notify If:
        + size delta between version is bigger than: 50%
        - time delta is irrelevant

TOKEN: TOKEN-UUID-B
  Name: content-example-b
  Write Only: false
  Keep N Versions: 3 (per file)
  Alert:
     Name: more than 1 day old config backup
     Matching on Regexp: "etc\.\.tar\.gz"
     UUID: NOTIFICATION-UUID
     Notify If:
        - size delta is irrelevant
        + last version is older than 1 days

  Alert:
     Name: file is 90% different
     Matching on Regexp: ".*"
     UUID: NOTIFICATION-UUID
     Notify If:
        + size delta between version is bigger than: 90%
        - time delta is irrelevant



## Create Token

curl -u your.email@example.com  -d '{
  "write_only":false,
  "keep_n_versions":7,
  "name": "example"
}' https://baxx.dev/protected/create/token


Write Only:
 tokens can only add but not download/list files (without password)

Keep #N Versions:
 How many versions per file to keep.  Useful for database or modified
 files archives like, e.g:

 mysqldump | curl --data-binary @- https://baxx.dev/io/$BAXX_TOKEN/mysql.gz


## Modify tokens

curl -u your.email@example.com \
 -d '{"write_only":false,"token":"TOKEN-UUID","name":"example"}' \
 https://baxx.dev/protected/change/token


## Delete tokens

curl -u your.email@example.com -d '{"token": "TOKEN-UUID"}' \
 https://baxx.dev/protected/delete/token

this will delete:
  * the token
  * all the files in it
  * all notifications attached to it


██╗    ██╗ ██████╗
██║   ██╔╝██╔═══██╗
██║  ██╔╝ ██║   ██║
██║ ██╔╝  ██║   ██║
██║██╔╝   ╚██████╔╝
╚═╝╚═╝     ╚═════╝

## File Upload

cat path/to/file | encrypt | curl --data-binary @- \
 https://baxx.dev/io/$BAXX_TOKEN/path/to/file

or (no encryption, strongly discouraged)
curl -T path/to/file https://baxx.dev/io/$BAXX_TOKEN/path/to/file

Same filepath can have up to #N Versions depending on the token
configuration.

## File Download

Download the last uploaded version of a file at specific path

curl https://baxx.dev/io/$BAXX_TOKEN/path/to/file > file

## File Delete

Delete single file:
curl -XDELETE https://baxx.dev/io/$BAXX_TOKEN/path/to/file

Delete all files in a directory, but not the subdirectories:
curl -d '{"force":true}' https://baxx.dev/io/$BAXX_TOKEN/path

## List Files

curl https://baxx.dev/ls/$BAXX_TOKEN/path/to

Lists files in path LIKE /path/to%
use '?format=json' if you want json back by default it prints human
readable text

## Write Only Tokens

Write Only tokens require BasicAuth.
The idea is that you can put them in in-secure places and not worry
about someone reading your data if they get stolen.

## Using WriteOnly tokens to access files:

* Download
curl -u your.email@example.com https://baxx.dev/io/$TOKEN/path/to/file

* Delete
curl -u your.email@example.com -XDELETE https://baxx.dev/io/$TOKEN/path/to/file

* List
curl -u your.email@example.com https://baxx.dev/ls/$TOKEN/path/

███████╗██╗   ██╗███╗   ██╗ ██████╗
██╔════╝╚██╗ ██╔╝████╗  ██║██╔════╝
███████╗ ╚████╔╝ ██╔██╗ ██║██║
╚════██║  ╚██╔╝  ██║╚██╗██║██║
███████║   ██║   ██║ ╚████║╚██████╗
╚══════╝   ╚═╝   ╚═╝  ╚═══╝ ╚═════╝

## GET: https://baxx.dev/sync/sha256/$BAXX_TOKEN/$sha

Returns non 200 status code if the sha does not exist
it is meant to be used with 'curl -f', which makes curl exit with non
zero in case of failure:

$sha is sha256 sum (shasum -a 256 file | cut -f 1 -d ' ')

check if sha exists, and upload if it doesnt
 curl -f https://baxx.dev/sync/sha256/$BAXX_TOKEN/$sha  || \
 curl -f -T $i https://baxx.dev/io/$BAXX_TOKEN/$i

## POST: https://baxx.dev/sync/sha256/$BAXX_TOKEN

This endpoint takes the multiple lines of shasum output
and returns only the lines that are not found, example input:

2997f66d71b5c0f2f396872536beed30835add1e1de8740b3136c9d550b1eb7c  a
8719d1dc6f98ebb5c04f8c1768342e865156b1582806b6c7d26e3fbdc99b8762  b
8d0a34b05558ad54c4a5949cc42636165b6449cf3324406d62e923bc060478dc  c
c7c2c1d3c83afbc522ae08779cd661546e578b2dfc6a398467d293bd63e03290  d


if you have already uploaded the file c it will return

2997f66d71b5c0f2f396872536beed30835add1e1de8740b3136c9d550b1eb7c  a
8719d1dc6f98ebb5c04f8c1768342e865156b1582806b6c7d26e3fbdc99b8762  b
c7c2c1d3c83afbc522ae08779cd661546e578b2dfc6a398467d293bd63e03290  d

it is very handy for rsync like uploads:
find | xargs shasum | curl diff | curl upload

example:
 find . -type f \
  | xargs -P4 -I '{}' \
    shasum -a 256 {} \
  | curl -s --data-binary @- https://baxx.dev/sync/sha256/$BAXX_TOKEN \
  | awk '{ print $2 }' \
  | xargs -P4 -I '{}' \
    curl -s -T {} https://baxx.dev/io/$BAXX_TOKEN/backup/{}

it is *VERY* important to curl to /sync/sha256 with --data-binary
otherwise curl is in ascii mode and does *not* send the new lines, and
only the first line is checked.

This is super annoying, and I am sure someone will lose backups
because of this, and there is nothing I can do about it.

This small script will find all files, then compute the shasums in
parallel check the diff with what is uploaded on baxx and upload only
the missing ones




███╗   ██╗ ██████╗ ████████╗██╗███████╗██╗   ██╗
████╗  ██║██╔═══██╗╚══██╔══╝██║██╔════╝╚██╗ ██╔╝
██╔██╗ ██║██║   ██║   ██║   ██║█████╗   ╚████╔╝
██║╚██╗██║██║   ██║   ██║   ██║██╔══╝    ╚██╔╝
██║ ╚████║╚██████╔╝   ██║   ██║██║        ██║
╚═╝  ╚═══╝ ╚═════╝    ╚═╝   ╚═╝╚═╝        ╚═╝

## Create Notification

curl -u your.email@example.com  -d '{
  "name":"example name",
  "token":"TOKEN-UUID",
  "regexp":".*",
  "age_days": 1,
  "size_delta_percent": 50
}' https://baxx.dev/protected/create/notification

* Name
  Human readable name that will be sent in the emails.

* age_days
  If the file has no new version in N days.

* size_delta_percent
  If the delta between the last version and previos version of the
  file is bigger than N.

  e.g.:
  previous version: example.txt - 500 bytes
   current version: example.txt - 10 bytes

  the alert will trigger and you will be notified

## Change Notification

curl -u your.email@example.com  -d '{
  "name":"example name",
  "notification_uuid":"NOTIFICATION-UUID",
  "regexp":".*",
  "age_days": 1,
  "size_delta_percent": 50
}' https://baxx.dev/protected/change/notification

## Delete Notification

curl -u your.email@example.com \
 -d '{"notification_uuid": "NOTIFICATION-UUID"}' \
 https://baxx.dev/protected/delete/notification
## List Notifications

To list your configured notifications you can use the /status endpoint:

curl -u your.email@example.com https://baxx.dev/protected/status



██████╗ ██████╗  ██████╗ ███████╗██╗██╗     ███████╗
██╔══██╗██╔══██╗██╔═══██╗██╔════╝██║██║     ██╔════╝
██████╔╝██████╔╝██║   ██║█████╗  ██║██║     █████╗
██╔═══╝ ██╔══██╗██║   ██║██╔══╝  ██║██║     ██╔══╝
██║     ██║  ██║╚██████╔╝██║     ██║███████╗███████╗
╚═╝     ╚═╝  ╚═╝ ╚═════╝ ╚═╝     ╚═╝╚══════╝╚══════╝

## Change Password

curl -u your.email@example.com -d'{"new_password": "donald mouse"}' \
 https://baxx.dev/protected/replace/password | json_pp

(use https://www.xkcd.com/936/)

## Change Email

curl -u your.email@example.com -d'{"new_email": "x@example.com"}' \
https://baxx.dev/protected/replace/email | json_pp

It will also send new verification email, you can also use the
replace/email endpoint to resend the verification email.

## User Status

curl -u your.email@example.com https://baxx.dev/protected/status

shows things like
 * is the email verified
 * is subscription active [ not done yet ]
 * current tokens
 * size used

## Register

This is the /register endpoint

curl -d '{"email":"your.email@example.com", "password":"mickey mouse"}' \
 https://baxx.dev/register


███████╗██╗  ██╗ █████╗ ███╗   ███╗██████╗ ██╗     ███████╗
██╔════╝╚██╗██╔╝██╔══██╗████╗ ████║██╔══██╗██║     ██╔════╝
█████╗   ╚███╔╝ ███████║██╔████╔██║██████╔╝██║     █████╗
██╔══╝   ██╔██╗ ██╔══██║██║╚██╔╝██║██╔═══╝ ██║     ██╔══╝
███████╗██╔╝ ██╗██║  ██║██║ ╚═╝ ██║██║     ███████╗███████╗
╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝     ╚═╝╚═╝     ╚══════╝╚══════╝




## upload everything from a directory

find . -type f -exec curl --data-binary @{}      \
              https://baxx.dev/io/$BAXX_TOKEN/{} \;

## upload in parallel

find . -type f | xargs -P 4 -I {} -- \
  curl -T {} https://baxx.dev/io/$BAXX_TOKEN/{}


## upload only the files that have difference in shasum

for i in $(find . -type f); do \
 echo -n "$i.."
 sha=$(shasum -a 256 $i | cut -f 1 -d ' ')
 (curl -s https://baxx.dev/sync/sha256/$BAXX_TOKEN/$sha -f && echo SKIP $i) || \
 (curl -T $i https://baxx.dev/io/$BAXX_TOKEN/$i -f)
done


## shell alias

### indentation is messed up  to fit 80 chars


export BAXX_TOKEN=...
baxx_put() {
 if [ $# -lt 2 ]; then
  echo "usage: $0 file dest [force]"
 else

  file=$1
  dest=$2
  force=${3:-noforce}

  sha=$(shasum -a 256 $file | cut -f 1 -d ' ')

  (curl -s https://baxx.dev/sync/sha256/$BAXX_TOKEN/$sha -f >/dev/null 2>&1 \
   && [[ "$force" != "force" ]] \
   && echo SKIP $file .. already baxxed, use \"$0 $1 $2 force\" to force) || \
  curl -T $file https://baxx.dev/io/$BAXX_TOKEN/$dest

 fi
}


baxx_get() {
 if [ $# -ne 2 ]; then
  echo "usage: $0 file dest"
 else
  file=$1
  dest=$2
  curl https://baxx.dev/io/$BAXX_TOKEN/$file > $dest
fi
}


baxx_delete() {
 if [ $# -ne 1 ]; then
  echo "usage: $0 file"
 else
  file=$1
  curl -X DELETE https://baxx.dev/io/$BAXX_TOKEN/$file
 fi
}

baxx_rmdir() {
 if [ $# -ne 1 ]; then
  echo "usage: $0 path"
 else
  path=$1
  curl -d '{"force":true}' \
    -X DELETE https://baxx.dev/io/$BAXX_TOKEN/$path
 fi
}

baxx_rmrf() {
 if [ $# -ne 1 ]; then
  echo "usage: $0 path"
 else
  path=$1
  curl -d '{"force":true,"recursive":true}' \
    -X DELETE https://baxx.dev/io/$BAXX_TOKEN/$path
 fi
}


baxx_ls() {
 curl https://baxx.dev/ls/$BAXX_TOKEN/$*
}

baxx_sync() {
 if [ $# -ne 1 ]; then
  echo "usage: $0 path"
 else
  find $1 -type f \
  | xargs -P4 -I '{}' \
    shasum -a 256 {} \
  | curl -s --data-binary @- https://baxx.dev/sync/sha256/$BAXX_TOKEN \
  | awk '{ print $2 }' \
  | xargs -P4 -I '{}' \
    curl -s -T {} https://baxx.dev/io/$BAXX_TOKEN/backup/{}
 fi
}



check out https://github.com/jackdoe/baxx/tree/master/examples
for more examples

then simply do
% baxx_put example.txt /some/dir/example.txt
2918    Sun Mar 10 07:08:35 2019        /some/dir/example.txt@v2755

% baxx_get /some/dir/example.txt example.txt.dl

███████╗███╗   ██╗ ██████╗██████╗ ██╗   ██╗██████╗ ████████╗
██╔════╝████╗  ██║██╔════╝██╔══██╗╚██╗ ██╔╝██╔══██╗╚══██╔══╝
█████╗  ██╔██╗ ██║██║     ██████╔╝ ╚████╔╝ ██████╔╝   ██║
██╔══╝  ██║╚██╗██║██║     ██╔══██╗  ╚██╔╝  ██╔═══╝    ██║
███████╗██║ ╚████║╚██████╗██║  ██║   ██║   ██║        ██║
╚══════╝╚═╝  ╚═══╝ ╚═════╝╚═╝  ╚═╝   ╚═╝   ╚═╝        ╚═╝


WE DO NOT ENCRYPT YOUR DATA
WE DO NOT ENCRYPT YOUR DATA
WE DO NOT ENCRYPT YOUR DATA
WE DO NOT ENCRYPT YOUR DATA
WE DO NOT ENCRYPT YOUR DATA

(well.. well we do with a per-token key, but dont trust it)
Always use encryption when sending data.


██╗   ██╗██████╗ ██╗      ██████╗  █████╗ ██████╗
██║   ██║██╔══██╗██║     ██╔═══██╗██╔══██╗██╔══██╗
██║   ██║██████╔╝██║     ██║   ██║███████║██║  ██║
██║   ██║██╔═══╝ ██║     ██║   ██║██╔══██║██║  ██║
╚██████╔╝██║     ███████╗╚██████╔╝██║  ██║██████╔╝
 ╚═════╝ ╚═╝     ╚══════╝ ╚═════╝ ╚═╝  ╚═╝╚═════╝

███████╗███╗   ██╗ ██████╗██████╗ ██╗   ██╗██████╗ ████████╗███████╗██████╗
██╔════╝████╗  ██║██╔════╝██╔══██╗╚██╗ ██╔╝██╔══██╗╚══██╔══╝██╔════╝██╔══██╗
█████╗  ██╔██╗ ██║██║     ██████╔╝ ╚████╔╝ ██████╔╝   ██║   █████╗  ██║  ██║
██╔══╝  ██║╚██╗██║██║     ██╔══██╗  ╚██╔╝  ██╔═══╝    ██║   ██╔══╝  ██║  ██║
███████╗██║ ╚████║╚██████╗██║  ██║   ██║   ██║        ██║   ███████╗██████╔╝
╚══════╝╚═╝  ╚═══╝ ╚═════╝╚═╝  ╚═╝   ╚═╝   ╚═╝        ╚═╝   ╚══════╝╚═════╝





--
baxx.dev


```


# who watches the watchers

the current baxx infra progress is: (still not live)

2 machines, each running only docker and ssh

```
[ b.baxx.dev ]
* ssh
* docker
  + postgres-master
  + who watches the watchers [job]
  + run notification rules [job]
  + process email queue [job]
  + collect memory/disk/mdadam stats [privileged] [job] (priv because mdadm)
  + baxx-api
  + judoc [localhost]
  + scylla [privileged] (priv because of io tunning)

[ a.baxx.dev ]
* ssh
* docker
  + postgres-slave
  + nginx + letsencrypt
  + who watches the watchers [job]
  + process email queue [job]
  + collect memory/disk/mdadam stats [privileged] [job] (priv because mdadm)
  + baxx-api
  + judoc [localhost]
  + scylla [privileged] (priv because of io tunning)
```

as you can see both machines are in the scylla cluster, and both of
them are sending the notification emails (using select for update locks)

I have built quite simple yet effective monitoring system for baxx.

Deach process with [job] tag is something like:

```
for {
    work
    sleep X
}
```

What I did is:

```
setup("monitoring key", X+5)
for {
    work
    tick("monitoring key")
    sleep X
}
```
Then the 'who watches the watchers' programs check if "monitoring key"
is executed at within X+5 seconds per node(), and if not they send
slack message

The who watches the watchers then sends notifications (both watchers
send notifications on their own, so i receive the notification twice
but that is ok)

The watchers themselves also use the system, so if one of them dies,
the other one will send notification.

# testing

## shut down postgres
✓ * shutdown postgres and see if notifications are sent

## mdadm

✓ * make it fail
  mdadm -f /dev/md2 /dev/nvme1n1p3

✓ * wait for panic message

✓ * remove the disk
  mdadm --remove /dev/md2 /dev/nvme1n1p3

✓ * add the disk back
  mdadm --add /dev/md2 /dev/nvme1n1p3

✓ * wait to see it is acknowledged

works really nice

## test disk thresh

✓ * start the status tool with with 1% disk threshold
    and wait for alert

## test memory thresh

* start the status tool with with 1% memory threshold
  and wait for alert


## test health of baxx api

* query /status which should
  + query postgres
  + query judoc
