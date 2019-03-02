package help

import "fmt"

func GenericHelp() string {
	return `
Alpha:

  this is alpha, use at your own risk

What:

  baxx.dev is a dead simple backup as a service with unix philosophy
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

  alpha is just on digitalocean, but the beta will be on dedicated
  servers from hetzner.com

  Your data is compressed and encrypted on input, but it is compressed
  without signing, so attackers can flip arbitrary bits, but not make
  sense out of it. The purpose of the encryption is just in case someone
  manages to get a file from the disk. You should always send encrypted
  data.


Cost:

  the alpha costs 5E per month for few GB much more than you would buy
  google drive or dropbox or anything

  ultimately you should be able to get 5E per 1tb per month (maybe
  even less)


`

}

func AfterRegistration(secret, tokenrw, tokenwo string) string {
	return fmt.Sprintf(`
Secret : %s

ReadWrite Token: %s
WriteOnly Token: %s
(they will be sent to your email as well).

[at the moment the email sending is not implemented]
[so copy paste the screen; move fast break things]

Backup: 
 cat path/to/file | curl --data-binary @- \
 https://baxx.dev/v1/io/$SECRET/$TOKEN/path/to/file

Restore: 
 curl https://baxx.dev/v1/io/$SECRET/$TOKEN/path/to/file > file

Restore from WriteOnly token: 
 curl -u email \\
 https://baxx.dev/protected/v1/io/$SECRET/$TOKEN/path/to/file

You can create new tokens at:
 curl -u email -d '{"WriteOnly":false, "NumberOfArchives":7}' \
 -XPOST https://baxx.dev/protected/v1/create/token

WriteOnly: 
 tokens can only add but not get files (without password)
NumberOfArchives: 
 how many versions per file (with different sha256) to keep

Useful for things like:
 mysqldump | curl curl --data-binary @- \\
 https://baxx.dev/v1/io/$SECRET/$TOKEN/mysql.gz

Help: 
 curl https://baxx.dev/v1/help
 ssh help@baxx.dev
 email help@baxx.dev
`, secret, tokenrw, tokenwo)
}
