package util

import (
	"os/user"
	"io/ioutil"
	"strconv"
	"strings"
)

func startsWith(s string, prefix string) bool {
	return s[0:len(prefix)] == prefix
}

func ExpandPath(path string) string {
	usr, _ := user.Current()
	dir := usr.HomeDir
	if startsWith(path, "~/") {
		path = dir + "/" + path[2:]
	}
	return path
}

func ReadFileBytes(path string) []byte {
	body, err := ioutil.ReadFile(ExpandPath(path))
	if err != nil {
		panic(err)
	}
	return body
}

func ReadFile(path string) string {
	body, err := ioutil.ReadFile(ExpandPath(path))
	if err != nil {
		panic(err)
	}
	return string(body)
}

func ReadFileFloat64(path string) float64 {
	body := ReadFile(path)
	amount, err := strconv.ParseFloat(strings.TrimSpace(body), 64)
	if err != nil {
		panic(err)
	}
	return amount
}
