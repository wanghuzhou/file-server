package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func md5Str(byte []byte) string {
	h := md5.New()
	h.Write(byte)
	return hex.EncodeToString(h.Sum(nil))
}

func success(s map[string]string) string {
	result := ResultDTO{
		Code: 0,
		Msg:  "",
		Data: s,
	}
	str, err := json.Marshal(result)
	checkErr(err)
	return string(str)
}

func toJsonString(v interface{}) string {
	str, err := json.Marshal(v)
	checkErr(err)
	return string(str)
}
