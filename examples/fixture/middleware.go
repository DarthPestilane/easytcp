package fixture

import (
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/DarthPestilane/easytcp/router"
	"github.com/sirupsen/logrus"
	"runtime/debug"
)

func RecoverMiddleware(log *logrus.Logger) router.MiddlewareFunc {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(ctx *router.Context) (*packet.MessageEntry, error) {
			defer func() {
				if r := recover(); r != nil {
					log.WithField("sid", ctx.SessionID()).Errorf("PANIC | %+v | %s", r, debug.Stack())
				}
			}()
			return next(ctx)
		}
	}
}
