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
)

var wg sync.WaitGroup

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
	img, _, err := image.Decode(file)
	if err != nil {
		return Histo{"", nil}, err
	}
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

func computeHistograms(imagePaths []string, depth int, hChan chan Histo) {
	//defer fmt.Print("|| TEST ||")
	for _, imagePath := range imagePaths {
		val, err := computeHistogram(("/Users/voldischool/Documents/GO-Projects/Similarity Image Search/res/imageDataset2_15_20/" + imagePath), depth)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
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
	for i := range h2.H {
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
	//[]string{"res/queryImages/q00.jpg",
	//	"res/imageDataset2_15_20/"}
	args := os.Args

	args[1] = "res/queryImages/" + args[1]
	args[2] = "res/" + args[2] + "/"

	k := 20
	dataset := splitSlice(readFiles(args[2]), k)
	histogramChannel := make(chan Histo, len(dataset)*len(dataset[0]))

	for _, subSlice := range dataset {
		wg.Add(1)
		go func(sub []string) {
			defer wg.Done()
			computeHistograms(sub, 10, histogramChannel)
		}(subSlice)
	}

	wg.Wait()
	close(histogramChannel)

	queryImage, err := computeHistogram(args[1], 10)
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
		if pair.distanceToQuery > highest[smallerIndex].distanceToQuery {
			highest[smallerIndex] = pair
		}
	}

	for _, value := range highest {
		fmt.Print("\n || " + value.h.Name + " ||")
	}
	fmt.Print("|| DONE ||")
}
