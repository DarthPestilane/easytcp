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
		TimeFormat: "2006-01-02T15:04:05.000Z07:00",
	}
}

func (f *TextFormatter) formatLevel(level logrus.Level) string {
	levelTxt := fmt.Sprintf("%-7s", strings.ToUpper(level.String())) // align level
	if f.WithColor {
		// @see https://en.wikipedia.org/wiki/ANSI_escape_code for colors code
		var levelColor int
		switch level {
		case logrus.DebugLevel, logrus.TraceLevel:
			levelColor = 37 // gray
		case logrus.WarnLevel:
			levelColor = 33 // yellow
		case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
			levelColor = 31 // red
		default:
			levelColor = 34 // blue
		}
		levelTxt = fmt.Sprintf("\u001B[%dm%s", levelColor, levelTxt)
	}
	return levelTxt
}

func (f *TextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var msg string
	// format level
	msg += f.formatLevel(entry.Level)
	// format timestamp
	msg += fmt.Sprintf(" [%s]", entry.Time.Format(f.TimeFormat))
	// format scope
	if scope, _ := entry.Data["scope"].(string); scope != "" {
		msg += fmt.Sprintf(" [%s]", scope)
	}
	// append message
	if entry.Message != "" {
		msg += fmt.Sprintf(" %s", entry.Message)
	}
	if f.WithColor {
		msg += "\u001B[0m"
	}
	// end the message with \n
	msg += "\n"
	return []byte(msg), nil

}
