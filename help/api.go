package help

import (
	"fmt"
	. "github.com/jackdoe/baxx/common"
	"strings"
)

func Token(status *UserStatusOutput) string {
	tokens := []string{}
	for _, t := range status.Tokens {
		tokens = append(tokens, fmt.Sprintf("TOKEN: %s\n  Write Only: %t, Keep N Versions: %d\n", t.ID, t.WriteOnly, t.NumberOfArchives))
	}
	return fmt.Sprintf(`SECRET: %s

 This is your user secret, its random uuid, and you should
 try to keep it safe, but in case you publish it somewhere
 you can replace it:

 curl -u %s -XPOST \
  https://baxx.dev/protected/v1/replace/secret | json_pp

Tokens
 they are like backup namespaces, you can have the same
 file in different tokens and it wont conflict

Current Tokens:
%s

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
`, status.Secret, status.Email, strings.Join(tokens, "\n"), status.Email)
}

func Backup(email string) string {
	return fmt.Sprintf(`File Upload:
 cat path/to/file | curl --data-binary @- \
   https://baxx.dev/v1/io/$SECRET/$TOKEN/path/to/file

File Download:
 curl https://baxx.dev/v1/io/$SECRET/$TOKEN/path/to/file > file

File Delete:
 curl -XDELETE https://baxx.dev/v1/io/$SECRET/$TOKEN/path/to/file

List Files in path LIKE /path/to%:
 curl https://baxx.dev/v1/ls/$SECRET/$TOKEN/path/to/
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
   https://baxx.dev/protected/v1/ls/$SECRET/$TOKEN/path/to/
`, email, email, email, email)
}

func Register(email string) string {
	return fmt.Sprintf(`Register:
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
