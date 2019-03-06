package help

var GUI_INTRO = Parse(`Storage 10G
  Trial 1 Month 0.1E
  Subscription: 5E per Month
  Availability: ALPHA`)

var GUI_PASS_REQUIRED = Parse(`Password is required.

If you are not using a password manager
please use good passwords, such as: 

  'mickey mouse and metallica'

https://www.xkcd.com/936/`)

var GUI_EMAIL_REQUIRED = Parse(`Email is required.

It we will not send you any marketing messages,
it will be used just for business, such as:
 * sending notifications when backups are
   delayed, smaller than normal
 * payment is received
 * payment is not received`)
