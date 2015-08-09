package util

import (
	"os/user"
	"io/ioutil"
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
