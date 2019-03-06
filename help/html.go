package help

var HTML_VERIFICATION_OK = Parse(`Thanks!
The email is verified now!

You can check your account status at:

  curl -u {{.Email}} -XPOST https://baxx.dev/protected/status

`)

var HTML_LINK_EXPIRED = Parse(`Oops, verification link has expired!

You can generate new one with:

 curl -u {{.Email}} \
  -XPOST -d'{"new_email": "{{.Email}}"}' \
  https://baxx.dev/protected/replace/email

The verification links are valid for 24 hours,
You can check your account status at:

  curl -u {{.Email}} -XPOST https://baxx.dev/protected/status

If something is wrong, please contact me at help@baxx.dev.

Thanks!`)

var HTML_LINK_ERROR = Parse(`Oops, something went wrong!

{{.Error}}

please contact me at jack@baxx.dev if this persists!`)
