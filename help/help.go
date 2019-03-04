package help

import (
	"fmt"
	. "github.com/jackdoe/baxx/common"
	//	"github.com/jackdoe/baxx/user"
)

func Intro() string {
	return `baxx.dev - Unix Philosophy Backup Service

Storage 10G
Trial 1 Month 0.1E
Subscription: 5E per Month
Availability: ALPHA
  Here be Dragons! Data can be lost.
`
}

func EmailAfterRegistration(status *UserStatusOutput) string {
	return fmt.Sprintf(`

%s
Subscription URL (redirects to paypal.com):

  https://baxx.dev/v1/sub/%s

Subscribe first in order to use the service.

## Secrets And Tokens

%s

## Storing and Loading files

%s

## Profile management

%s

## Terms and Conditions

%s

## GDPR

%s

Help:
 curl https://baxx.dev/v1/help [ not ready yet ]
 ssh help@baxx.dev [ not ready yet ]
 email jack@baxx.dev
`, Intro(), status.PaymentID, Token(status), Backup(status.Email), Register(status.Email), TermsAndConditions(), GDPR())
}
