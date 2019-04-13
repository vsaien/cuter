package stringx

import (
	"errors"

	"github.com/vsaien/cuter/lib/lang"
)

var (
	ErrInvalidStartPosition = errors.New("start position is invalid")
	ErrInvalidStopPosition  = errors.New("stop position is invalid")
)

func Contains(list []string, str string) bool {
	for _, each := range list {
		if each == str {
			return true
		}
	}

	return false
}

func Reverse(s string) string {
	runes := []rune(s)

	for from, to := 0, len(runes)-1; from < to; from, to = from+1, to-1 {
		runes[from], runes[to] = runes[to], runes[from]
	}

	return string(runes)
}

func Substr(str string, start int, end int) (string, error) {
	rs := []rune(str)
	length := len(rs)

	if start < 0 || start > length {
		return "", ErrInvalidStartPosition
	}

	if end < 0 || end > length {
		return "", ErrInvalidStopPosition
	}

	return string(rs[start:end]), nil
}

func TakeOne(valid, or string) string {
	if len(valid) > 0 {
		return valid
	} else {
		return or
	}
}

func Union(first, second []string) []string {
	set := make(map[string]lang.PlaceholderType)

	for _, each := range first {
		set[each] = lang.Placeholder
	}
	for _, each := range second {
		set[each] = lang.Placeholder
	}

	merged := make([]string, 0, len(set))
	for k := range set {
		merged = append(merged, k)
	}

	return merged
}

func Remove(strings []string, str string) []string {
	var out []string
	for _, v := range strings {
		if v != str {
			out = append(out, v)
		}
	}
	return out
}
