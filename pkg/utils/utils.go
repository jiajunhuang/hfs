package utils

import (
	"encoding/json"
)

func ToJSONString(c interface{}) (string, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
