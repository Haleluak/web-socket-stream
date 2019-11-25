package util

import (
	"encoding/json"
	"strings"
)

func Beautify(object interface{}) string {
	s, _ := json.MarshalIndent(object, "", strings.Repeat(" ", 4))
	return string(s)
}
