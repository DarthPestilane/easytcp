package fixture

import (
	"github.com/DarthPestilane/easytcp"
	"github.com/DarthPestilane/easytcp/message"
	"github.com/sirupsen/logrus"
	"runtime/debug"
)

func RecoverMiddleware(log *logrus.Logger) easytcp.MiddlewareFunc {
	return func(next easytcp.HandlerFunc) easytcp.HandlerFunc {
		return func(c *easytcp.Context) (*message.Entry, error) {
			defer func() {
				if r := recover(); r != nil {
					log.WithField("sid", c.Session().ID()).Errorf("PANIC | %+v | %s", r, debug.Stack())
				}
			}()
			return next(c)
		}
	}
}
