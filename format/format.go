package format

import "time"

var DateTimeFormat = time.RFC1123

func TimeToGMT(t time.Time) string {
	t_str := t.Format(DateTimeFormat)

	return t_str[:len(t_str)-3] + "GMT"
}
