package fixture

import (
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/DarthPestilane/easytcp/router"
	"github.com/DarthPestilane/easytcp/session"
	"github.com/sirupsen/logrus"
	"runtime/debug"
)

func RecoverMiddleware(log *logrus.Logger) router.MiddlewareFunc {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(s session.Session, req *packet.Request) (*packet.Response, error) {
			defer func() {
				if r := recover(); r != nil {
					log.Errorf("PANIC | %+v | %s", r, debug.Stack())
				}
			}()
			return next(s, req)
		}
	}
}
