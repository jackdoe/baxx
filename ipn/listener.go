package ipn

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
)

//LiveIPNEndpoint contains the notification verification URL
const LiveIPNEndpoint = "https://www.paypal.com/cgi-bin/webscr"

//SandboxIPNEndpoint is the Sandbox notification verification URL
const SandboxIPNEndpoint = "https://ipnpb.sandbox.paypal.com/cgi-bin/webscr"

func getEndpoint(testIPN bool) string {
	if testIPN {
		return SandboxIPNEndpoint
	}
	return LiveIPNEndpoint
}

//Listener creates a PayPal listener.
//if err is set in cb, PayPal will resend the notification at some future point.
func Listener(g *gin.Engine, path string, cb func(c *gin.Context, err error, body string, n *Notification) error) {
	g.POST(path, func(c *gin.Context) {
		w := c.Writer
		r := c.Request

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			cb(c, errors.Wrap(err, "failed to read body"), "", nil)
			return
		}

		form, err := url.ParseQuery(string(body))
		if err != nil {
			cb(c, errors.Wrap(err, "failed to parse query"), "", nil)
			return
		}

		notification := ReadNotification(form)

		log.Debugf("paypal: form: %s, parsed: %+v\n", body, notification)

		body = append([]byte("cmd=_notify-validate&"), body...)

		resp, err := http.Post(getEndpoint(notification.TestIPN), r.Header.Get("Content-Type"), bytes.NewReader(body))
		if err != nil {
			cb(c, errors.Wrap(err, "failed to create post verification req"), "", nil)
			return
		}
		defer resp.Body.Close()

		verifyStatus, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			cb(c, errors.Wrap(err, "failed to read verification response"), "", nil)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if string(verifyStatus) != "VERIFIED" {
			cb(c, errors.Errorf("unexpected verify status %q", verifyStatus), "", nil)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// notification confirmed here

		// tell PayPal to not send more notificatins
		err = cb(c, nil, string(body), notification)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	})
}
