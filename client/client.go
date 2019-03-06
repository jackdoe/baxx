package client

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	. "github.com/jackdoe/baxx/common"
	"io/ioutil"
	"net/http"
	"strings"
)

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

type Client struct {
	h      *http.Client
	base   string
	status chan string
}

func NewClient(h *http.Client, base string, status chan string) *Client {
	if h == nil {
		h = &http.Client{}
	}
	return &Client{h: h, base: base, status: status}
}

func (c *Client) query(path string, user string, pass string, req interface{}, dest interface{}) error {
	comm := func(msg ...string) {
		if c.status != nil {
			c.status <- fmt.Sprintf("query %s %s", path, strings.Join(msg, " "))
		}
	}
	comm("...")
	encoded, err := json.Marshal(req)
	if err != nil {
		comm("error", err.Error())
		return err
	}

	url := c.base
	if strings.HasSuffix(url, "/") {
		url = fmt.Sprintf("%s%s", url, path)
	} else {
		url = fmt.Sprintf("%s/%s", url, path)
	}

	httpreq, err := http.NewRequest("POST", url, bytes.NewBuffer(encoded))
	if user != "" {
		httpreq.SetBasicAuth(user, pass)
	}

	if err != nil {
		comm("error", err.Error())
		return err
	}
	resp, err := c.h.Do(httpreq)
	if err != nil {
		comm("error", err.Error())
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		comm("error", err.Error())
		return err
	}
	if resp.StatusCode != http.StatusOK {
		qe := &QueryError{}
		err = json.Unmarshal(body, qe)
		if err != nil {
			comm("error", err.Error())
			return err
		}
		err = errors.New(qe.Error)
		comm("error", err.Error())
		return err
	} else {
		err = json.Unmarshal(body, dest)
		if err != nil {
			comm("error", err.Error())
			return err
		}
		comm("...", "done")
	}
	return nil
}

func (c *Client) Register(input *CreateUserInput) (*UserStatusOutput, error) {
	out := &UserStatusOutput{}
	err := c.query("register", "", "", input, out)
	return out, err
}

func (c *Client) Status(input *CreateUserInput) (*UserStatusOutput, error) {
	out := &UserStatusOutput{}
	err := c.query("protected/status", input.Email, input.Password, map[string]string{}, out)
	return out, err
}

func (c *Client) ReplaceEmail(input *CreateUserInput, newEmail string) (*UserStatusOutput, error) {
	out := &UserStatusOutput{}
	err := c.query("protected/replace/email", input.Email, input.Password, ChangeEmailInput{NewEmail: newEmail}, out)
	return out, err
}
