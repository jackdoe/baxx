# baxx.dev

check it out `ssh register@ui.baxx.dev`

[ work in progress ]

* https://baxx.dev/help
* TODO.txt
* infra and pricing.txt
* stat (disk usage, mem, mdadm) https://baxx.dev/stat
# backup service
(also i am learning how to build a product without a website haha)

# screenshots

┌───────────────────────────────────────────────┐
│                                               │
│ ██████╗  █████╗ ██╗  ██╗██╗  ██╗              │
│ ██╔══██╗██╔══██╗╚██╗██╔╝╚██╗██╔╝              │
│ ██████╔╝███████║ ╚███╔╝  ╚███╔╝               │
│ ██╔══██╗██╔══██║ ██╔██╗  ██╔██╗               │
│ ██████╔╝██║  ██║██╔╝ ██╗██╔╝ ██╗              │
│ ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝              │
│                                               │
│ Storage 10G                                   │
│   Trial 1 Month 0.1 EUR                       │
│   Subscription: 5 EUR per Month               │
│   Availability: ALPHA                         │
│                                               │
│ Contact Us:                                   │
│  * Slack         https://baxx.dev/join/slack  │
│  * Google Groups https://baxx.dev/join/groups │
│                                               │
│ E-mail                                        │
│                                               │
│ Password                                      │
│                                               │
│ Confirm Password                              │
│                                               │
│                                               │
│ Registering means you agree with              │
│ the terms of service!                         │
│                                               │
│              [Register]  [Login]              │
│                                               │
│   [Help]  [What/Why/How]  [Terms Of Service]  │
│                                               │
│                     [Quit]                    │
└───────────────────────────────────────────────┘



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


# who watches the watchers

the current baxx infra progress is

2 machines, each running only docker and ssh

[ b.baxx.dev ]
* ssh
* docker
  + postgres-master
  + nginx + letsencrypt
  + who watches the watchers [👹job]
  + run notification rules [👹job]
  + process email queue [👹job]
  + collect memory/disk/mdadam stats [privileged] [👹job] (priv because mdadm)
  + baxx-api
  + judoc [localhost]
  + scylla [privileged] (priv because of io tunning)

[ a.baxx.dev ]
* ssh
* docker
  + postgres-slave
  + nginx + letsencrypt
  + who watches the watchers [👹job]
  + process email queue [👹job]
  + collect memory/disk/mdadam stats [privileged] [👹job] (priv because mdadm)
  + baxx-api
  + judoc [localhost]
  + scylla [privileged] (priv because of io tunning)

as you can see both machines are in the scylla cluster, and both of
them are sending the notification emails (using select for update locks)
and only one of them is running the notification rules.

I have built quite simple yet effective monitoring system for baxx.

Each process with [👹job] tag is something like:
(using 👹 because of daemon)

  for {
      work
      sleep X
  }

What I did is:

  setup("monitoring key", X+5)
  for {
      work
      tick("monitoring key")
      sleep X
  }


Then the 'who watches the watchers' programs check if "monitoring key"
is executed at within X+5 seconds per node(), and if not they send
slack message

The 'who watches the watchers' then sends notifications (both watchers
send notifications on their own, so i receive the notification twice
but that is ok)

The watchers themselves also use the system, so if one of them dies,
the other one will send notification.

# testing

all the ✓ checks are tested (manually) and the alerts are performing
really good

## shut down postgres
* ✓ shutdown postgres and see if notifications are sent

## shut down one machine
* ✓ aa.baxx.dev
* ✓ bb.baxx.dev

## mdadm

* ✓ make it fail
  mdadm -f /dev/md2 /dev/nvme1n1p3

* ✓ wait for panic message

* ✓ remove the disk
  mdadm --remove /dev/md2 /dev/nvme1n1p3

* ✓ add the disk back
  mdadm --add /dev/md2 /dev/nvme1n1p3

* ✓ wait to see it is acknowledged

works really nice

## test disk thresh

* ✓ start the status tool with with 1% disk threshold
    and wait for alert

## test memory thresh

* start the status tool with with 1% memory threshold
  and wait for alert


## test health of baxx api

* query /status which should
  + query postgres
  + query judoc
