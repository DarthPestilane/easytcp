package message

import (
	"fmt"
	"regexp"
	"strconv"
)

var headRegexp = regexp.MustCompile(`^\[([0-9a-zA-Z\-_><.*?/()]+)\](\d+)\|$`)

type Head struct {
	RoutePath string
	Length    int
}

func ExtractHead(head []byte) (*Head, error) {
	if !headRegexp.Match(head) {
		return nil, fmt.Errorf("invalid message head")
	}
	matches := headRegexp.FindStringSubmatch(string(head))
	path := matches[1]
	length, err := strconv.Atoi(matches[2])
	if err != nil {
		return nil, err
	}
	return &Head{RoutePath: path, Length: length}, nil
}

func AddHead(routePath string, body []byte) (message []byte) {
	head := fmt.Sprintf("[%s]%d|", routePath, len(body))
	message = append([]byte(head), body...)
	return
}
