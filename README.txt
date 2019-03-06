# baxx.dev

check it out ssh register@baxx.dev

[ work in progress ]


* backup service
(also i am learning how to build a product without a website haha)

# screenshots

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
│   Trial 1 Month 0.1E                       │
│   Subscription: 5E per Month               │
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


Post registration screen:                                                                                                                                                                                              
                                                                                                                                                                                               
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
                                                                                                                                                                                              

# API

┌baxx.dev - HELP──────────────────────────────────────────────────────────────────────────────┐
│                                                                                             │
│ Hi,                                                                                         │
│                                                                                             │
│ The service I offer is still in Alpha stage, but I really appreciate                        │
│ the support.                                                                                │
│                                                                                             │
│ # Subscription                                                                              │
│                                                                                             │
│ ## Plan (only one for now):                                                                 │
│                                                                                             │
│ Storage 10G                                                                                 │
│ Trial 1 Month 0.1E                                                                          │
│ Subscription: 5E per Month                                                                  │
│ Availability: ALPHA                                                                         │
│                                                                                             │
│ Here be Dragons! Data can be lost!                                                          │
│                                                                                             │
│ ## Subscribe                                                                                │
│                                                                                             │
│ In order to use baxx.dev you need a subscription,                                           │
│ At the moment I support only paypal.com, please visit:                                      │
│                                                                                             │
│ https://baxx.dev/sub/WILL-BE-IN-YOUR-EMAIL                                                  │
│ To be redirected to paypal.com.                                                             │
│                                                                                             │
│ ## Verify your email                                                                        │
│                                                                                             │
│ Email verification is also required, you should've received the                             │
│ verification link in another email.                                                         │
│                                                                                             │
│ Thanks again!                                                                               │
│                                                                                             │
│ # Tokens                                                                                    │
│                                                                                             │
│ Tokens are like backup namespaces, you can have the same file in                            │
│ different tokens and it won't conflict.                                                     │
│                                                                                             │
│ ## Current Tokens:                                                                          │
│                                                                                             │
│                                                                                             │
│   TOKEN: TOKEN-UUID-A                                                                       │
│   Write Only: true                                                                          │
│   Keep N Versions 3                                                                         │
│                                                                                             │
│                                                                                             │
│   TOKEN: TOKEN-UUID-B                                                                       │
│   Write Only: false                                                                         │
│   Keep N Versions 3                                                                         │
│                                                                                             │
│                                                                                             │
│ ## Create New Tokens:                                                                       │
│                                                                                             │
│ curl -u your.email@example.com \                                                            │
│  -d '{"write_only":false, "keep_n_versions":7}' \                                           │
│  https://baxx.dev/protected/create/token                                                    │
│                                                                                             │
│ Write Only:                                                                                 │
│  tokens can only add but not get files (without password)                                   │
│                                                                                             │
│ Keep #N Versions:                                                                           │
│  How many versions per file (with different sha256) to keep.  Useful                        │
│  for database or modified files archives like, e.g:                                         │
│                                                                                             │
│  mysqldump | curl --data-binary @- \                                                        │
│   https://baxx.dev/io/$TOKEN/mysql.gz                                                       │
│                                                                                             │
│ ## Delete tokens                                                                            │
│                                                                                             │
│ curl -u your.email@example.com -d '{"uuid": "TOKEN-UUID"}' \                                │
│  https://baxx.dev/protected/delete/token                                                    │
│                                                                                             │
│ this will delete the token and all the files in it                                          │
│                                                                                             │
│ # File operations                                                                           │
│                                                                                             │
│ ## File Upload:                                                                             │
│                                                                                             │
│ cat path/to/file | curl --data-binary @- \                                                  │
│  https://baxx.dev/io/$TOKEN/path/to/file                                                    │
│                                                                                             │
│ Same filepath can have up to #N Versions depending on the token                             │
│ configuration.                                                                              │
│                                                                                             │
│ Uploading the same sha256 resulting in reusing existing version and                         │
│ also does not consume quota.                                                                │
│                                                                                             │
│ ## File Download:                                                                           │
│                                                                                             │
│ curl https://baxx.dev/io/$TOKEN/path/to/file > file                                         │
│                                                                                             │
│ Downloads the last upload version                                                           │
│                                                                                             │
│ ## File Delete:                                                                             │
│                                                                                             │
│ curl -XDELETE https://baxx.dev/io/$TOKEN/path/to/file                                       │
│                                                                                             │
│ deletes all versions of a file                                                              │
│                                                                                             │
│ ## List Files in path LIKE /path/to%:                                                       │
│                                                                                             │
│ curl https://baxx.dev/ls/$TOKEN/path/to                                                     │
│                                                                                             │
│ use -H "Accept: application/json" if you want json back by default it                       │
│ prints human readable text                                                                  │
│                                                                                             │
│ ## WriteOnly Tokens                                                                         │
│                                                                                             │
│ Write Only tokens require BasicAuth and /protected prefix.                                  │
│ The idea is that you can put them in in-secure places and not worry                         │
│ about someone reading your data if they get stolen.                                         │
│                                                                                             │
│ ## Download from WriteOnly token:                                                           │
│                                                                                             │
│ curl -u your.email@example.com \                                                            │
│  https://baxx.dev/protected/io/$TOKEN/path/to/file                                          │
│                                                                                             │
│ ## Delete with WriteOnly token:                                                             │
│                                                                                             │
│ curl -u your.email@example.com -XDELETE \                                                   │
│  https://baxx.dev/io/$TOKEN/path/to/file                                                    │
│                                                                                             │
│ ## List with WriteOnly token:                                                               │
│                                                                                             │
│ curl -u your.email@example.com \                                                            │
│  https://baxx.dev/protected/ls/$TOKEN/path/to/                                              │
│                                                                                             │
│                                                                                             │
│ # Profile Management                                                                        │
│                                                                                             │
│ ## Register:                                                                                │
│                                                                                             │
│ curl -d '{"email":"your.email@example.com", "password":"mickey mouse"}' \                   │
│  https://baxx.dev/register | json_pp                                                        │
│                                                                                             │
│ ## Change Password                                                                          │
│                                                                                             │
│ curl -u your.email@example.com -d'{"new_password": "donald mouse"}' \                       │
│  https://baxx.dev/protected/replace/password | json_pp                                      │
│                                                                                             │
│ (use https://www.xkcd.com/936/)                                                             │
│                                                                                             │
│ ## Change Email                                                                             │
│                                                                                             │
│ curl -u your.email@example.com -d'{"new_email": "x@example.com"}' \                         │
│  https://baxx.dev/protected/replace/email | json_pp                                         │
│                                                                                             │
│ It will also send new verification email, you can also use the                              │
│ replace/email endpoint to resend the verification email.                                    │
│                                                                                             │
│ ## User Status                                                                              │
│                                                                                             │
│ curl -u your.email@example.com -XPOST https://baxx.dev/protected/status                     │
│                                                                                             │
│ shows things like                                                                           │
│  * is the email verified                                                                    │
│  * is subscription active [ not done yet ]                                                  │
│  * current tokens                                                                           │
│  * size used                                                                                │
│                                                                                             │
│ # Encryption                                                                                │
│                                                                                             │
│ Your data is compressed and encrypted when received, the encryption                         │
│ key is auto generated uuid, and the purpose of the encryption is                            │
│ simply to obscure the data in case the machines are hacked, hacker                          │
│ will have to also get access to the database as well.                                       │
│                                                                                             │
│ Anyway, dont trust it and use encryption when uploading.                                    │
│                                                                                             │
│ --                                                                                          │
│ baxx.dev                                                                                    │
│[█Back ]                                                                                     │
└─────────────────────────────────────────────────────────────────────────────────────────────┘




