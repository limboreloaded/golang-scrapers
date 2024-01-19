package utils

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func ModifyFile(file *os.File, key string, value string, selector string) {
	body, err := ioutil.ReadFile(file.Name())
	if err != nil {
		log.Fatalf("unable to read file: %v", err)
	}

	str := strings.Replace(string(body), fmt.Sprintf(selector, key), fmt.Sprintf(selector, value), 1)

	_ = file.Truncate(0)
	_, err = file.Seek(0, 0)

	if err != nil {
		log.Fatal(err)
	}

	file.Write([]byte(str))
}

func ChangeKeysByPath(m map[string]string, json *os.File, md *os.File, csv *os.File) {
	for key := range m {
		ModifyFile(json, key, m[key], `"%s"`)
		ModifyFile(md, key, m[key], "%s:")
		ModifyFile(csv, key, m[key], "%s")
	}
}
