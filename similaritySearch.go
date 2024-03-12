package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
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

// splitSlice splits a slice into K slices
func splitSlice(slice []string, k int) [][]string {
	n := len(slice)
	if k <= 0 || n == 0 {
		return nil
	}

	result := make([][]string, k)
	avgSize := n / k
	remainder := n % k
	start := 0
	end := 0

	for i := 0; i < k; i++ {
		end += avgSize
		if i < remainder {
			end++
		}
		if end > n {
			end = n
		}
		result[i] = slice[start:end]
		start = end
	}

	return result
}

func computeHistograms(imagePaths []string, depth int, hChan chan<- Histo) {
	defer close(hChan)

	for _, imagePath := range imagePaths {
		val, err := computeHistogram(imagePath, depth)
		if err != nil {
			continue
		}
		hChan <- val
	}
}

func main() {
	args := os.Args

	wg := sync.WaitGroup{}

	histogramChannel := make(chan Histo)
	k := 1
	dataset := splitSlice(readFiles(args[1]), k)

	for _, subSlice := range dataset {
		wg.Add(1)
		go computeHistograms(subSlice, 10, histogramChannel)
	}

	queryImage, err := computeHistogram(args[0], 10)

	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

}
