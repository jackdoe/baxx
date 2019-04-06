package message

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

type Slack struct {
	Title    string   `json:"title"`
	Text     string   `json:"text"`
	MrkdwnIn []string `json:"mrkdwn_in"`
}

func Hostname() string {
	name, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	return name
}

func SendSlack(webhook string, title string, body string) error {
	log.Infof("sending slack message %s %s", title, body)
	encoded, err := json.Marshal(Slack{
		Title:    title,
		Text:     Hostname() + ": " + title + "\n" + body,
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
			return
		}
		if resp.StatusCode != 200 {
			wait <- errors.New(resp.Status)
		}
		wait <- nil
	}()
	select {
	case err := <-wait:
		return err
	case <-time.After(3 * time.Second):
		return errors.New("slack request timed out")
	}
}

func MustHaveMonitoring() {
	hook := os.Getenv("BAXX_SLACK_MONITORING")
	if hook == "" {
		log.Panic("must have BAXX_SLACK_MONITORING")
	}
}

func MustHavePanic() {
	hook := os.Getenv("BAXX_SLACK_PANIC")
	if hook == "" {
		log.Panic("must have BAXX_SLACK_PANIC")
	}
}

func MustHaveInfo() {
	hook := os.Getenv("BAXX_SLACK_INFO")
	if hook == "" {
		log.Panic("must have BAXX_SLACK_INFO")
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

func SendSlackMonitoring(title, body string) {
	hook := os.Getenv("BAXX_SLACK_MONITORING")
	if hook != "" {
		err := SendSlack(hook, title, body)
		if err != nil {
			log.Warnf("error sending to slack: %s", err.Error())
		}
	} else {
		log.Warnf("not sending to slack: %s/%s", title, body)
	}
}

func SendSlackInfo(title, body string) {
	hook := os.Getenv("BAXX_SLACK_INFO")
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
		m := ""
		if l, ok := r.(*logrus.Entry); ok {
			s, err := l.String()
			if err != nil {
				m = fmt.Sprintf("%s\n```%s```\n```%s```", r, os.Args, stack)
			} else {
				m = fmt.Sprintf("%s\n```%s```\n```%s```", s, os.Args, stack)
			}
		} else {
			m = fmt.Sprintf("%s\n```%s```\n```%s```", r, os.Args, stack)
		}
		SendSlackDefault(topic, m)
		panic(r)
	}
}
