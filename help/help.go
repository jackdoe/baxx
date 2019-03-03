package help

import "fmt"

func TermsAndConditions() string {
	return `
By purchasing, registering for, or using the “Baxx” services (the
  “Services”) you ("referred in the document as "you", "customer",
  "subscriber", "client") enter into a contract with Borislav Nikolov
  Amsterdam, The Netherlands (also referred in the document as "baxx",
  "we"),and you accept and agree to the following terms (the
  “Contract”). The Contract shall apply to the supply of the Services,
  use of the Services after purchase or after registering for limited
  free use where such offer has been made available.

Services:
  We provide the service of storing your data, with specific retention
  rate, depending on the the offer you have registered the scope of
  this service might vary.

Acceptable Conduct:
  You are responsible for the actions of all users of your account and
  any data that is created, stored, displayed by, or transmitted by
  your account while using Baxx. You will not engage in any activity
  that interferes with or disrupts Baxx's services or networks
  connected to Baxx.

Contract Duration
  You agree that any malicious activities are considered prohibited
  usage and will result in immediate account suspension or
  cancellation without a refund and the possibility that we will
  impose fees; and/or pursue civil remedies without providing advance
  notice.

  You agree that Baxx shall be permitted to charge your credit card on
  a monthly basis in advance of providing services. Payment is due
  every month. Service may be interrupted on accounts that reach 1
  month past due.

Backup
  Subscriber is solely responsible for the preservation of
  Subscriber's data which Subscriber saves onto its Baxx account (the
  “Data”). Even with respect to Data as to which Subscriber contracts
  for backup services provided by Baxx, We shall have no
  responsibility to preserve Data. We shall have no liability for any
  Data that may be lost.

  The data must sent *encrypted* to to Baxx, and can not be read by
  Baxx employees or third parties.


Use of the service is at your own risk.

THE SERVICE IS PROVIDED "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES,
INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF
MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL I BE LIABLE FOR ANY DIRECT, INDIRECT,
INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING,
BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS
OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR
TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE
USE OF THIS SERVICE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH
DAMAGE.
`
}

func License() string {
	return `
Copyright (C) 2018 Borislav Nikolov

This software is provided 'as-is', without any express or implied
warranty. In no event will the authors be held liable for any damages
arising from the use of this software.

Permission is granted to anyone to use this software for any purpose,
including commercial applications, and to alter it and redistribute it
freely, subject to the following restrictions:

1. The origin of this software must not be misrepresented; you must not
   claim that you wrote the original software. If you use this software
   in a product, an acknowledgment in the product documentation would be
   appreciated but is not required.
2. Altered source versions must be plainly marked as such, and must not be
   misrepresented as being the original software.
3. This notice may not be removed or altered from any source distribution.

Borislav Nikolov
jack@sofialondonmoskva.com
`
}

func GDPR() string {
	return `
We are not sharing the data with anyone for no purposes what so ever.
We are keeping logs of IP adress uploading/downloading files,and the
paypal payment notifications for starting/ending the subscription.
`
}

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

func Backup(email string) string {
	return fmt.Sprintf(`
File Upload:
 cat path/to/file | curl --data-binary @- \
   https://baxx.dev/v1/io/$SECRET/$TOKEN/path/to/file

File Download:
 curl https://baxx.dev/v1/io/$SECRET/$TOKEN/path/to/file > file

File Delete:
 curl -XDELETE https://baxx.dev/v1/io/$SECRET/$TOKEN/path/to/file

List Files in path LIKE /path/to%:
 curl https://baxx.dev/v1/dir/$SECRET/$TOKEN/path/to/
 use -H "Accept: application/json" if you want json back
 by default it prints human readable text

WriteOnly tokens require BasicAuth and /protected prefix.

Download from WriteOnly token:
 curl -u %s \
   https://baxx.dev/protected/v1/io/$SECRET/$TOKEN/path/to/file

Delete with WriteOnly token:
 curl -u %s -XDELETE \
   https://baxx.dev/v1/io/$SECRET/$TOKEN/path/to/file

List with WriteOnly token:
 curl -u %s \
   https://baxx.dev/protected/v1/dir/$SECRET/$TOKEN/path/to/
`, email, email, email, email)
}

func Token(email string) string {
	return fmt.Sprintf(`
Secret
 This is your user secret, its random uuid, and you should
 try to keep it safe, but in case you publish it somewhere
 you can replace it:

 curl -u %s -XPOST \
  https://baxx.dev/protected/v1/replace/secret | json_pp

Tokens
 they are like backup namespaces, you can have the same
 file in different tokens and it wont conflict


Create New Tokens:
 curl -u %s -d '{"write_only":false, "keep_n_versions":7}' \
   https://baxx.dev/protected/v1/create/token

write_only:
 tokens can only add but not get files (without password)
keep_n_versions:
 How many versions per file (with different sha256) to keep.
 Useful for database or modified files archives like, e.g:
 mysqldump | curl curl --data-binary @- \
  https://baxx.dev/v1/io/$SECRET/$TOKEN/mysql.gz
`, email, email)
}

func Register(email string) string {
	return fmt.Sprintf(`

Register:
 curl -d '{"email":"%s", "password":"mickey mouse"}' \
  https://baxx.dev/v1/register | json_pp

Change Password
 curl -u %s -d'{"new_password": "donald mouse"}' \
  https://baxx.dev/protected/v1/replace/password | json_pp

Change Email
 curl -u %s -d'{"new_email": "x@example.com"}' \
  https://baxx.dev/protected/v1/replace/email | json_pp

 It will also send new verification email, you can
 also use the replace/email endpoint to resend the
 verification email.

User Status
 curl -u %s -XPOST https://baxx.dev/protected/v1/status

 Check the user status such as:
  * is the email verified
  * is subscription active [ not done yet ]

`, email, email, email, email)
}
func Intro() string {
	return `baxx.dev - Unix Philosophy Backup Service
Storage 10G
Trial 1 Month 0.1E
Subscription: 5E per Month
Availability: ALPHA
  Here be Dragons! Data can be lost.
  Refunds in case everything burns down.
`
}

func AfterRegistration(payment, email, secret, tokenrw, tokenwo string) string {
	return fmt.Sprintf(`

%s
Subscription URL (redirects to paypal.com):

  https://baxx.dev/v1/sub/%s

Please subscribe first in order to use the service.


Secret : %s

ReadWrite Token: %s
WriteOnly Token: %s
(they will be sent to your email as well).

%s

%s

%s


Terms of service

%s

GDPR

%s

Help:
 curl https://baxx.dev/v1/help [ not ready yet ]
 ssh help@baxx.dev [ not ready yet ]
 email jack@baxx.dev
`, Intro(), payment, secret, tokenrw, tokenwo, Backup(email), Token(email), Register(email), TermsAndConditions(), GDPR())
}
