package main

import (
	"fmt"

	"net/http/httputil"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/jackdoe/baxx/common"
	log "github.com/sirupsen/logrus"
)

func SlackRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				//debug.PrintStack()
				httprequest, _ := httputil.DumpRequest(c.Request, false)
				stack := debug.Stack()
				log.Warnf("[Recovery] panic recovered:\n%s\n%s\n%s", string(httprequest), err, stack)

				if CONFIG.SlackWebHook != "" {
					m := fmt.Sprintf("%s%s ```%s```", httprequest, err, stack)
					err := common.SendSlack(CONFIG.SlackWebHook, "panic", m)
					if err != nil {
						log.Warnf("error sending to slack: %s", err.Error())
					}
				}
				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	}
}
