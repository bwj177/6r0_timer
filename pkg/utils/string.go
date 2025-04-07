package utils

import "encoding/json"

func GetJsonStr(src any) string {
	dst, _ := json.Marshal(src)
	return string(dst)
}
