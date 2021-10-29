package fixture

import (
	"github.com/DarthPestilane/easytcp"
	"github.com/sirupsen/logrus"
	"runtime/debug"
)

func RecoverMiddleware(log *logrus.Logger) easytcp.MiddlewareFunc {
	return func(next easytcp.HandlerFunc) easytcp.HandlerFunc {
		return func(c easytcp.Context) {
			defer func() {
				if r := recover(); r != nil {
					log.WithField("sid", c.Session().ID()).Errorf("PANIC | %+v | %s", r, debug.Stack())
				}
			}()
			next(c)
		}
	}
}
