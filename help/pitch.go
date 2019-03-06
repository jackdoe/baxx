package help

var PITCH = Parse(`# baxx.dev is a simple backup service with unix philosophy in mind

  tackling what is fundamental problem with backups:
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

## Price

 Storage 10G
 Trial 1 Month 0.1E (10 euro cents)
 Subscription: 5E per Month

## How:

 The Alpha is just on digitalocean, but the beta will be on dedicated
 servers from hetzner.com

 Your data is compressed and encrypted on input, but it is compressed
 without signing, so attackers can flip arbitrary bits, but not make
 sense out of it. The purpose of the encryption is just in case
 someone manages to get a file from the disk. You should always send
 encrypted data.`)
