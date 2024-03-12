package main

import (
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
)

type Histo struct {
	Name string
	H    []float32
}

// adapted from: first example at pkg.go.dev/image
func computeHistogram(imagePath string, depth int) (Histo, error) {
	// Open the JPEG file
	file, err := os.Open(imagePath)
	if err != nil {
		return Histo{"", nil}, err
	}
	defer file.Close()

	// Decode the JPEG image
	file.Seek(0, 0)
	img, _, err := image.Decode(file)
	fmt.Print("|| USED ||")
	if err != nil {
		return Histo{"", nil}, err
	}
	fmt.Print("|| USED ||")
	// Get the dimensions of the image
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	result := make([]float32, depth)
	rgbMax := make([]float32, 10)
	// Scaning the RGB values for the image
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

func computeHistograms(imagePaths []string, depth int, wg *sync.WaitGroup, hChan chan<- Histo) {
	defer wg.Done()

	for _, imagePath := range imagePaths {
		val, err := computeHistogram(imagePath, depth)
		if err != nil {
			continue
		}
		hChan <- val
	}
}

func min(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func compareHistograms(h1 Histo, h2 Histo) (result float32) {
	result = 0.0
	for i, _ := range h2.H {
		result += min(h1.H[i], h2.H[i])
	}
	return result
}

type Pair struct {
	h               Histo
	distanceToQuery float32
}

func minPair(slice []Pair) (index int) {
	index = 0
	for i, value := range slice {
		if value.distanceToQuery < slice[index].distanceToQuery {
			index = i
		}
	}
	return index
}

func main() {
	fmt.Print("|| STARTING ||")
	args := []string{"/Users/voldischool/Documents/GO-Projects/Similarity Image Search/res/queryImages/q00.jpg",
		"/Users/voldischool/Documents/GO-Projects/Similarity Image Search/res/imageDataset2_15_20/"}

	wg := sync.WaitGroup{}

	k := 1
	dataset := splitSlice(readFiles(args[1]), k)
	histogramChannel := make(chan Histo, len(dataset[0])*len(dataset))

	for _, subSlice := range dataset {
		wg.Add(1)
		go computeHistograms(subSlice, 10, &wg, histogramChannel)
	}
	close(histogramChannel)

	queryImage, err := computeHistogram(args[0], 10)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	wg.Wait()

	highest := make([]Pair, 0, 5)
	for value := range histogramChannel {
		pair := Pair{value, compareHistograms(queryImage, value)}
		if len(highest) != 5 {
			highest = append(highest, pair)
			continue
		}

		smallerIndex := minPair(highest)
		if pair.distanceToQuery > highest[smallerIndex].distanceToQuery {
			highest[smallerIndex] = pair
		}
	}

	for _, value := range highest {
		fmt.Print("|| " + value.h.Name + " ||")
	}

}
