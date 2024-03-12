package main

import (
	"io/ioutil"
	"log"
	"strings"
)

func readFiles(directoryPath string) (filenames []string) {
	files, err := ioutil.ReadDir(directoryPath)
	if err != nil {
		log.Fatal(err)
	}

	// get the list of jpg files
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".jpg") {
			filenames = append(filenames, file.Name())
		}
	}
	return filenames
}

func main() {

}
