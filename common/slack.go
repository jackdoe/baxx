package common

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

type Slack struct {
	Title    string   `json:"title"`
	Text     string   `json:"text"`
	MrkdwnIn []string `json:"mrkdwn_in"`
}

func SendSlack(webhook string, title string, body string) error {
	encoded, err := json.Marshal(Slack{
		Title:    title,
		Text:     body,
		MrkdwnIn: []string{"text"},
	})
	if err != nil {
		return err
	}

	resp, err := http.Post(webhook, "application/json", bytes.NewReader(encoded))
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errors.New(resp.Status)
	}
	return nil
}
