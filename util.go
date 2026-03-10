package mailglyph

import "strconv"

func intToString(v int) string {
	return strconv.Itoa(v)
}

func boolToString(v bool) string {
	return strconv.FormatBool(v)
}
