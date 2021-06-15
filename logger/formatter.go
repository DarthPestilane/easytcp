package logger

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
)

type TextFormatter struct {
	WithColor  bool
	TimeFormat string
}

func NewTextFormatter() *TextFormatter {
	return &TextFormatter{
		WithColor:  true,
		TimeFormat: "2006-01-02T15:04:05.000000Z07:00",
	}
}

func (f *TextFormatter) formatLevel(level logrus.Level) string {
	return fmt.Sprintf("%-7s", strings.ToUpper(level.String())) // align level
}

func (f *TextFormatter) paintColor(level logrus.Level, msg string) string {
	if f.WithColor {
		// @see https://en.wikipedia.org/wiki/ANSI_escape_code for colors code
		var color int
		switch level {
		case logrus.TraceLevel:
			color = 90 // dark gray
		case logrus.DebugLevel:
			color = 37 // light gray
		case logrus.WarnLevel:
			color = 33 // yellow
		case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
			color = 31 // red
		default:
			color = 34 // blue
		}
		msg = fmt.Sprintf("\033[%dm%s\033[0m", color, msg)
	}
	return msg
}

func (f *TextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var msg string
	// format level
	msg += f.formatLevel(entry.Level)
	// format timestamp
	msg += fmt.Sprintf(" %s |", entry.Time.Format(f.TimeFormat))
	// format session id
	if sid, _ := entry.Data["sid"].(string); sid != "" {
		msg += fmt.Sprintf(" %s |", sid)
	}
	// format scope
	if scope, _ := entry.Data["scope"].(string); scope != "" {
		msg += fmt.Sprintf(" %s |", scope)
	}
	// append message
	if entry.Message != "" {
		msg += fmt.Sprintf(" %s", entry.Message)
	}
	// paint color
	msg = f.paintColor(entry.Level, msg)
	// end the message with \n
	msg += "\n"
	return []byte(msg), nil
}
