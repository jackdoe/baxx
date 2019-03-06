package help

var PITCH = Parse(`# baxx.dev is a simple backup service with unix philosophy in mind

Tackling the fundamental problems with backups:
  * anomaly detection
  * notifications
  * alerts
  * durability
  * watching the watchers

In 20 years of experience I have used many backup solutions, from one
line scripts to complicated systems, and in almost all cases when I
needed the data it was either corrupt or 0 bytes, or was missing.

## Alpha:

baxx.dev is still alpha, use at your own risk
no guarnatees about the data [yet]

## Why charging during alpha:

Because I want to see if someone really cares about this.
Lets work together to make usable backup service!

## Encryption

Your data is compressed and encrypted when received, the encryption
key is auto generated uuid, and the purpose of the encryption is
simply to obscure the data in case the machines are hacked, hacker
will have to also get access to the database as well.

Anyway, dont trust it and use encryption when sending data.

## End goal

* zero configuration notifications
  using active learning (or machine teaching as they call it now)

* ask when uncertain
  when some anomaly is detected, keep the files and wait for manual
  confirmation (via email)

* easy configuration
  easy and intuitive rules that can be shared

## Open Source

The whole thing will be open sourced (soon)

`)
