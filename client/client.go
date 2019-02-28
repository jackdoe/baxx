package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	. "github.com/jackdoe/baxx/common"
	. "github.com/jackdoe/baxx/user"
	"io/ioutil"
	"net/http"
	"strings"
)

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

func (c *Client) query(path string, req interface{}, dest interface{}) error {
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

	resp, err := c.h.Post(url, "application/json", bytes.NewBuffer(encoded))
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

func (c *Client) Register(input *CreateUserInput) (*User, error) {
	user := &User{}
	err := c.query("v1/register", input, user)
	return user, err
}
