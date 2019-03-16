package notification

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	log "github.com/sirupsen/logrus"
)

type Slack struct {
	Title    string   `json:"title"`
	Text     string   `json:"text"`
	MrkdwnIn []string `json:"mrkdwn_in"`
}

func SendSlack(webhook string, title string, body string) error {
	log.Infof("sending slack message %s %s", title, body)
	encoded, err := json.Marshal(Slack{
		Title:    title,
		Text:     body,
		MrkdwnIn: []string{"text"},
	})
	if err != nil {
		return err
	}
	wait := make(chan error, 1)
	go func() {
		resp, err := http.Post(webhook, "application/json", bytes.NewReader(encoded))
		if err != nil {
			wait <- err
		}
		if resp.StatusCode != 200 {
			wait <- errors.New(resp.Status)
		}
	}()
	select {
	case err := <-wait:
		return err
	case <-time.After(3 * time.Second):
		return errors.New("slack request timed out")
	}
}

func SendSlackDefault(title, body string) {
	hook := os.Getenv("BAXX_SLACK_PANIC")
	if hook != "" {
		err := SendSlack(hook, title, body)
		if err != nil {
			log.Warnf("error sending to slack: %s", err.Error())
		}
	} else {
		log.Warnf("not sending to slack: %s/%s", title, body)
	}
}

func SlackPanic(topic string) {
	if r := recover(); r != nil {
		stack := debug.Stack()
		m := fmt.Sprintf("%s: %s ```%s```", topic, r, stack)

		SendSlackDefault(topic, m)
		panic(r)
	}
}
