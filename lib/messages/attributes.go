package messages

import (
	"fmt"
	"strconv"
)

func GetAttrInt(m map[string]interface{}, key string) (int, error) {
	value, err := GetAttrString(m, key)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(value)
}

func GetAttrString(m map[string]interface{}, key string) (string, error) {
	value, ok := m[key]
	if !ok {
		return "", fmt.Errorf("%s is not set", key)
	}

	str, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("%s is not a string", key)
	}

	return str, nil
}

func GetAttrUint64(m map[string]interface{}, key string) (uint64, error) {
	value, err := GetAttrString(m, key)
	if err != nil {
		return 0, err
	}

	return strconv.ParseUint(value, 10, 64)
}
