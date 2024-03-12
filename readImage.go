package main

// import (
// 	"fmt"
// 	"image"
// 	_ "image/jpeg"
// 	"os"
// )

// type Histo struct {
// 	Name string
// 	H    []float32
// }

// // adapted from: first example at pkg.go.dev/image
// func computeHistogram(imagePath string, depth int) (Histo, error) {
// 	// Open the JPEG file
// 	file, err := os.Open(imagePath)
// 	if err != nil {
// 		return Histo{"", nil}, err
// 	}
// 	defer file.Close()

// 	// Decode the JPEG image
// 	img, _, err := image.Decode(file)
// 	if err != nil {
// 		return Histo{"", nil}, err
// 	}

// 	// Get the dimensions of the image
// 	bounds := img.Bounds()
// 	width, height := bounds.Max.X, bounds.Max.Y
// 	result := make([]float32, depth)
// 	rgbMax := make([]float32, 10)
// 	// Scaning the RGB values for the image
// 	for y := 0; y < height; y++ {
// 		for x := 0; x < width; x++ {

// 			// Convert the pixel to RGBA
// 			red, green, blue, _ := img.At(x, y).RGBA()
// 			// A color's RGBA method returns values in the range [0, 65535].
// 			// Shifting by 8 reduces this to the range [0, 255].
// 			red >>= 8
// 			blue >>= 8
// 			green >>= 8

// 			//Data for normalization
// 			rgbMax[0] += float32(red)
// 			rgbMax[1] += float32(blue)
// 			rgbMax[2] += float32(green)

// 			result = append(result, float32(red), float32(blue), float32(green))

// 			fmt.Printf("Pixel at (%d, %d): R=%d, G=%d, B=%d\n", x, y, red, green, blue)
// 		}
// 	}

// 	//Normalizing the Histogram
// 	for i, value := range result {
// 		result[i] = value / rgbMax[i%3]
// 	}

// 	h := Histo{imagePath, result}
// 	return h, nil
// }

// func main() {
// 	// read the image name from command line
// 	args := "/Users/voldischool/Documents/GO-Projects/Similarity Image Search/res/queryImages/q00.jpg"

// 	// Call the function to display RGB values of some pixels
// 	_, err := computeHistogram(args, 10)
// 	if err != nil {
// 		fmt.Printf("Error: %s\n", err)
// 		return
// 	}
// }
