package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

var wg sync.WaitGroup

type Histo struct {
	Name string
	H    []float32
}

// adapted from: first example at pkg.go.dev/image
// Computes the histogram for the given image. It is assumed that the image is JPEG encoded and has the given depth
//
// @param imagePath - The path to the JPG image
// @param depth - The depth of the histogram to be computed 0
func computeHistogram(imagePath string, depth int) (Histo, error) {
	// Open the JPEG file
	file, err := os.Open(imagePath)
	// Return a Histo object with the error if any
	if err != nil {
		return Histo{"", nil}, err
	}
	defer file.Close()

	// Decode the JPEG image
	img, _, err := image.Decode(file)
	// Return a Histo object with the error if any
	if err != nil {
		return Histo{"", nil}, err
	}
	// Get the dimensions of the image
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	result := make([]float32, depth)
	rgbMax := make([]float32, 10)
	// Scaning the RGB values for the image
	// Convert the image to RGBA values.
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {

			// Convert the pixel to RGBA
			red, green, blue, _ := img.At(x, y).RGBA()
			// A color's RGBA method returns values in the range [0, 65535].
			// Shifting by 8 reduces this to the range [0, 255].
			red >>= 8
			blue >>= 8
			green >>= 8

			//Data for normalization
			rgbMax[0] += float32(red)
			rgbMax[1] += float32(blue)
			rgbMax[2] += float32(green)

			result = append(result, float32(red), float32(blue), float32(green))
		}
	}

	//Normalizing the Histogram
	for i, value := range result {
		result[i] = value / rgbMax[i%3]
	}

	h := Histo{imagePath, result}
	return h, nil
}

// readFiles reads all jpg files in a directory and returns a list of filenames. If an error occurs it will log the error and exit the program
//
// @param directoryPath - the path to the directory
// @param filenames - the list of jpg
func readFiles(directoryPath string) (filenames []string) {
	files, err := ioutil.ReadDir(directoryPath)
	// This method is called when the error occurs.
	if err != nil {
		log.Fatal(err)
	}

	// Add. jpg files to the list of filenames.
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".jpg") {
			filenames = append(filenames, file.Name())
		}
	}
	return filenames
}

// Splits a slice into k sub slices. This is useful for splitting large slices in a multi dimensional array such as a csv. Rows
//
// @param slice - the slice to be split
// @param k - the number of sub elements to be returned. If k is negative or zero the result will be
func splitSlice(slice []string, k int) [][]string {
	n := len(slice)
	// Returns nil if k or n is negative.
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
		// Set the end of the list
		if end > n {
			end = n
		}
		result[i] = slice[start:end]
		start = end
	}

	return result
}

// Computes histogram for each image. This is a wrapper around GOPATH computeHistogram which takes a channel to send histograms to
//
// @param imagePaths - List of paths to images
// @param depth - Depth of histogram to compute 0 for no depth
// @param hChan - channel used to send histograms to the main thread
func computeHistograms(imagePaths []string, depth int, hChan chan Histo) {
	//defer fmt.Print("|| TEST ||")
	for _, imagePath := range imagePaths {
		val, err := computeHistogram(("res/imageDataset2_15_20/" + imagePath), depth)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			continue
		}
		hChan <- val
	}
}

// min returns the smaller of two values. The values are clamped to the range 0 1. This is useful for sorting floating point numbers in a way that makes it easier to compare values without losing precision.
//
// @param a - First value to compare. Must be non negative.
//
// @return The smaller of a and b or a if a b is greater than b in which case the result is a
func min(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

// Compare two histograms and return the sum of the minimum values. This is used to determine the percentage to which a histogram is the same as another histogram.
//
// @param h1 - First histogram to compare.
// @param h2 - Second histogram to compare.
// @param result
func compareHistograms(h1 Histo, h2 Histo) (result float32) {
	result = 0.0
	for i := range h2.H {
		result += min(h1.H[i], h2.H[i])
	}

	return result
}

type Pair struct {
	h          Histo
	similarity float32
}

// Returns the index of the Pair that maximises the similarity. This is used to sort Pairs in ascending order
//
// @param slice - slice of Pair to search
// @param index - index of Pair that minimizes the distanceTo
func minPair(slice []Pair) (index int) {
	index = 0
	for i, value := range slice {
		if value.similarity < slice[index].similarity {
			index = i
		}
	}
	return index
}

// Main function for similarity search. It takes two arguments queryImage and a dataset directory
// in this form : go run similaritySearch.go q00.jpg imageDataset2_15_20
func main() {
	fmt.Print("|| STARTING ||")

	args := os.Args

	args[1] = "res/queryImages/" + args[1]
	args[2] = "res/" + args[2] + "/"

	startTime := time.Now()

	k := 16
	dataset := splitSlice(readFiles(args[2]), k)
	histogramChannel := make(chan Histo, len(dataset)*len(dataset[0]))

	// computeHistograms computes histograms for each subSlice in the dataset.
	//Also the loop that we will get the execution time with different K
	for _, subSlice := range dataset {
		wg.Add(1)
		go func(sub []string) {
			defer wg.Done()
			computeHistograms(sub, 10, histogramChannel)
		}(subSlice)
	}

	wg.Wait()
	close(histogramChannel)

	endTime := time.Now()

	queryImage, err := computeHistogram(args[1], 10)
	// Print error message if any.
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	highest := make([]Pair, 0, 5)
	for value := range histogramChannel {
		pair := Pair{value, compareHistograms(queryImage, value)}
		if len(highest) != 5 {
			highest = append(highest, pair)
			continue
		}

		smallerIndex := minPair(highest)
		if pair.similarity > highest[smallerIndex].similarity {
			highest[smallerIndex] = pair
		}
	}

	// Prints the highest highest value.
	for _, value := range highest {
		fmt.Print("\n || " + value.h.Name + " ||")
	}
	fmt.Print("|| DONE || \n")
	executionTime := endTime.Sub(startTime)
	fmt.Print("|| Execution time : ", executionTime, " ||")
}
