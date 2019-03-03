package ipn

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/schema"
	"net/url"
	"time"
)

var decoder = schema.NewDecoder()

type Time struct {
	Time *time.Time
}

// PaymentStatus represents the status of a payment
type PaymentStatus string

// Payment statuses
var (
	PaymentStatusCanceledReversal PaymentStatus = "Canceled_Reversal"
	PaymentStatusCompleted        PaymentStatus = "Completed"
	PaymentStatusCreated          PaymentStatus = "Created"
	PaymentStatusDenied           PaymentStatus = "Denied"
	PaymentStatusExpired          PaymentStatus = "Expired"
	PaymentStatusFailed           PaymentStatus = "Failed"
	PaymentStatusPending          PaymentStatus = "Pending"
	PaymentStatusReversed         PaymentStatus = "Reversed"
	PaymentStatusProcessed        PaymentStatus = "Processed"
	PaymentStatusVoided           PaymentStatus = "Voided"
)

// PaymentType represents the type of a payment
type PaymentType string

// Payment Types
var (
	PaymentTypeEcheck  PaymentType = "echeck"
	PaymentTypeInstant PaymentType = "instant"
)

// PendingReason represents the reason the payment is pending
type PendingReason string

// Pending reasons
var (
	PendingReasonAddress             PendingReason = "address"
	PendingReasonAuthorization       PendingReason = "authorization"
	PendingReasonDelayedDisbursement PendingReason = "delayed_disbursement"
	PendingReasonEcheck              PendingReason = "echeck"
	PendingResasonIntl               PendingReason = "intl"
	PendingReasonMultiCurrency       PendingReason = "multi_currency"
	PendingReasonOrder               PendingReason = "order"
	PendingReasonPaymentReview       PendingReason = "paymentreview"
	PendingReasonRegulatoryReview    PendingReason = "regulatory_review"
	PendingReasonUnilateral          PendingReason = "unilateral"
	PendingReasonUpgrade             PendingReason = "upgrade"
	PendingReasonVerify              PendingReason = "verify"
	PendingReasonOther               PendingReason = "other"
)

// Notification is sent from PayPal to our application.
// See https://developer.paypal.com/docs/classic/ipn/integration-guide/IPNandPDTVariables for more info
type Notification struct {
	TxnType          string `schema:"txn_type"`
	TxnID            string `schema:"txn_id"`
	Business         string `schema:"business"`
	Custom           string `schema:"custom"`
	ParentTxnID      string `schema:"parent_txn_id"`
	ReceiptID        string `schema:"receipt_id"`
	RecieverEmail    string `schema:"receiver_email"`
	RecieverID       string `schema:"receiver_id"`
	Resend           bool   `schema:"resend"`
	ResidenceCountry string `schema:"residence_country"`
	TestIPN          bool   `schema:"test_ipn"`
	ItemName         string `schema:"item_name"`
	ItemNumber       string `schema:"item_number"`

	//Buyer address information
	AddressCountry     string `schema:"address_country"`
	AddressCity        string `schema:"address_city"`
	AddressCountryCode string `schema:"address_country_code"`
	AddressName        string `schema:"address_name"`
	AddressState       string `schema:"address_state"`
	AddressStatus      string `schema:"address_status"`
	AddressStreet      string `schema:"address_street"`
	AddressZip         string `schema:"address_zip"`

	//Misc buyer info
	ContactPhone      string `schema:"contact_phone"`
	FirstName         string `schema:"first_name"`
	LastName          string `schema:"last_name"`
	PayerBusinessName string `schema:"payer_business_name"`
	PayerEmail        string `schema:"payer_email"`
	PayerID           string `schema:"payer_id"`
	PayerStatus       string `schema:"payer_status"`

	AuthAmount string `schema:"auth_amount"`
	AuthExpire string `schema:"auth_exp"`
	AuthIfD    string `schema:"auth_id"`
	AuthStatus string `schema:"auth_status"`
	Invoice    string `schema:"invoice"`

	//Payment amount
	Currency string  `schema:"mc_currency"`
	Fee      float64 `schema:"mc_fee"`
	Gross    float64 `schema:"mc_gross"`

	PaymentDate   Time          `schema:"payment_date"`
	PaymentStatus PaymentStatus `schema:"payment_status"`
	PaymentType   PaymentType   `schema:"payment_type"`
	PendingReason PendingReason `schema:"pending_reason"`

	//ReasonCode is populated if the payment is negative
	ReasonCode string `schema:"reason_code"`

	Memo string `schema:"memo"`
}

//CustomerInfo returns a nicely formatted customer info string
func (n *Notification) JSON() (string, error) {
	b, err := json.Marshal(n)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (n *Notification) CustomerInfo() string {
	const form = `%v %v
%v
%v, %v, %v, %v %v
%v`
	return fmt.Sprintf(form, n.FirstName, n.LastName, n.PayerEmail, n.AddressStreet, n.AddressCity, n.AddressState, n.AddressZip, n.AddressCountry, n.PayerStatus)
}

//ReadNotification reads a notification from an //IPN request
func ReadNotification(vals url.Values) *Notification {
	n := &Notification{}
	decoder.Decode(n, vals) //errors due to missing fields in struct
	return n
}

const timeLayout = "15:04:05 Jan 02, 2006 MST"

func (t *Time) UnmarshalText(text []byte) (err error) {
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return err
	}
	tmp, err := time.ParseInLocation(timeLayout, string(text), loc)
	if err != nil {
		return err
	}
	t.Time = &tmp
	return nil
}
