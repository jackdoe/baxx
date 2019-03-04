package help

func GenericHelp() string {
	return `
Alpha:

  this is alpha, use at your own risk
  not everything is ready yet
  no guarnatees about the data yet

Why charging during alpha:

  Because I want to see if someone really cares about this.
  Lets work together to make usable backup as a service!


What:

  baxx.dev is a simple backup as a service with unix philosophy
  in mind

  tackling what is fundamental problem with backups:
   * anomaly detection
   * notifications
   * alerts

  and not storage availability, if someone tells you 99.9999%
  availability is important for backups, they are lying to you.

Why:

  because I want to explore the idea of building a service without a
  website, share it with my friends and have fun

How:

  The Alpha is just on digitalocean, but the beta will be on dedicated
  servers from hetzner.com

  Your data is compressed and encrypted on input, but it is compressed
  without signing, so attackers can flip arbitrary bits, but not make
  sense out of it. The purpose of the encryption is just in case
  someone manages to get a file from the disk. You should always send
  encrypted data.


Cost:

  The alpha costs 5E per month for 10 GB much more expensive than you
  would buy google drive or dropbox or anything (for now).

  Ultimately you should be able to get 5E per 1TB per month (maybe
  even less).

`
}
