package myimage

import (
	"fmt"
	"github.com/longears/sortpixels/mycolor"
	"github.com/longears/sortpixels/utils"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"math"
	"math/rand"
	"os"
)

// The random number generator
var RNG *rand.Rand

func init() {
	RNG = rand.New(rand.NewSource(99))
}

//================================================================================
// IMAGE

type MyImage struct {
	xres   int
	yres   int
	pixels [][]*mycolor.MyColor // 2d array, [x][y]
}

// Given a path to an image file on disk, return a MyImage struct.
func MakeMyImageFromPath(path string) *MyImage {
	// open file and decode image
	fmt.Println("  reading and decoding image")
	file, err := os.Open(path)
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		// couldn't decode; probably the file is not actually an image
		fmt.Println("  can't decode image.")
	}

	// convert to MyImage
	fmt.Println("  converting to MyImage")
	myImage := &MyImage{}
	myImage.populateFromNativeImage(img)
	return myImage
}

// Init the MyImage pixel array, creating MyColor objects
// from the data in the given image (from the built-in image package).
// HSV is computed here also for each pixel.
func (i *MyImage) populateFromNativeImage(img image.Image) {
	i.xres = img.Bounds().Max.X
	i.yres = img.Bounds().Max.Y
	i.pixels = make([][]*mycolor.MyColor, i.xres)
	for x := 0; x < i.xres; x++ {
		i.pixels[x] = make([]*mycolor.MyColor, i.yres)
		for y := 0; y < i.yres; y++ {
			r, g, b, a := img.At(x, y).RGBA()
			c := &mycolor.MyColor{uint8(r / 256), uint8(g / 256), uint8(b / 256), uint8(a / 256), 0, 0, 0, 0}
			c.ComputeHSV()
			i.pixels[x][y] = c
		}
	}
}

func handleErr(err error) {
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}
}

func (i *MyImage) SaveAs(path string) {
	// convert back to built in image
	fmt.Println("  converting to built in image")
	destImg := i.ToBuiltInImage()

	// write output
	fmt.Println("  writing to", path)
	fo, err := os.Create(path)
	handleErr(err)
	defer func() {
		err := fo.Close()
		handleErr(err)
	}()
	png.Encode(fo, destImg)
}

// Create an image using the built-in image.RGBA type
// and copy our pixels into it.
func (i *MyImage) ToBuiltInImage() image.Image {
	destImg := image.NewRGBA(image.Rectangle{image.ZP, image.Point{i.xres, i.yres}})
	for x := 0; x < i.xres; x++ {
		for y := 0; y < i.yres; y++ {
			myColor := i.pixels[x][y]
			rgba := color.RGBA{uint8(myColor.R), uint8(myColor.G), uint8(myColor.B), uint8(myColor.A)}
			destImg.Set(x, y, rgba)
		}
	}
	return destImg
}

func (i *MyImage) String() string {
	return fmt.Sprintf("<image %v x %v>", i.xres, i.yres)
}

//================================================================================
// SORTING

// Read y coordinates over yChan and sort those rows.
// Send 1 to doneChan when each row is done.
// The image natively stores pixels in columns, not rows, so we
// have to copy the pixels into a temporary slice, sort it, then
// put it back.
func (i *MyImage) goSortRow(kind string, yChan chan int, doneChan chan int) {
	row := make([]*mycolor.MyColor, i.xres)
	for y := range yChan {
		// copy into temp slice
		// set sort value
		for x := 0; x < i.xres; x++ {
			row[x] = i.pixels[x][y]
			row[x].SetSortValue(kind, x)
		}
		// sort
		utils.SortF(
			len(row),
			func(a, b int) bool {
				return row[a].SortValue < row[b].SortValue
			},
			func(a, b int) {
				row[a], row[b] = row[b], row[a]
			})
		// copy back into main array
		for x := 0; x < i.xres; x++ {
			i.pixels[x][y] = row[x]
		}
		doneChan <- 1
	}
}

// Launch some threads to sort the rows of the image.
// Wait until complete, and kill the threads.
func (i *MyImage) SortRows(kind string, numThreads int) {
	yChan := make(chan int, i.yres+10)
	doneChan := make(chan int, i.yres+10)
	for threadNum := 0; threadNum < numThreads; threadNum++ {
		go i.goSortRow(kind, yChan, doneChan)
	}

	for y := 0; y < i.yres; y++ {
		yChan <- y
	}
	close(yChan)
	for y := 0; y < i.yres; y++ {
		<-doneChan
	}
}

// Read slices of MyColor pointers over toSortChan and sort them.
// Send 1 to doneChan when each row is done.
// The image natively stores pixels in colums so we don't need to
// create any temporary slices here.
func goSortMyColorSlice(kind string, toSortChan chan []*mycolor.MyColor, doneChan chan int) {
	for colorSlice := range toSortChan {
		// set sort value
		for ii, v := range colorSlice {
			v.SetSortValue(kind, ii)
		}
		// do actual sort
		utils.SortF(
			len(colorSlice),
			func(a, b int) bool {
				return colorSlice[a].SortValue < colorSlice[b].SortValue
			},
			func(a, b int) {
				colorSlice[a], colorSlice[b] = colorSlice[b], colorSlice[a]
			})
		doneChan <- 1
	}
}

// Launch some threads to sort the rows of the image.
// Wait until complete, and kill the threads.
func (i *MyImage) SortColumns(kind string, numThreads int) {
	toSortChan := make(chan []*mycolor.MyColor, i.xres+10)
	doneChan := make(chan int, i.xres+10)
	for threadNum := 0; threadNum < numThreads; threadNum++ {
		go goSortMyColorSlice(kind, toSortChan, doneChan)
	}

	for x := 0; x < i.xres; x++ {
		toSortChan <- i.pixels[x]
	}
	close(toSortChan)
	for x := 0; x < i.xres; x++ {
		<-doneChan
	}
}

//================================================================================
// CONGREGATE

type kernelElem struct {
	x        int
	y        int
	strength float64
}

func makeKernel(radius int) []kernelElem {
	// MAKE KERNEL
	// kernel is a list of x,y,strength
	kernel := make([]kernelElem, 1)
	for x := -radius; x <= radius; x++ {
		for y := -radius; y <= radius; y++ {
			if x == 0 && y == 0 {
				continue
			}
			dist := math.Hypot(float64(x), float64(y)) / float64(radius)
			strength := 1.0 / (dist*dist + 0.2)
			kernel = append(kernel, kernelElem{x, y, strength})
		}
	}
	return kernel
}

// Return a float between 0 and 1 indicating how similar the colors are.
func colorSimilarity(a *mycolor.MyColor, b *mycolor.MyColor) float64 {
	return 1.0 - (math.Abs(float64(a.H-b.H))+math.Abs(float64(a.S-b.S))+math.Abs(float64(a.V-b.V)))/3.0
}

// Given a color, a coordinate, and a kernel, compute the fitness
// of that color at that location (compared to its neighbors)
func (i *MyImage) pixelFitness(color *mycolor.MyColor, x int, y int, kernel []kernelElem) float64 {
	totalStrength := float64(0)
	totalFitness := float64(0)
	for _, elem := range kernel {
		thisX := elem.x + x
		thisY := elem.y + y
		if thisX < 0 || thisX >= i.xres {
			continue
		}
		if thisY < 0 || thisY >= i.yres {
			continue
		}
		otherColor := i.pixels[thisX][thisY]
		totalFitness += colorSimilarity(color, otherColor) * elem.strength
		totalStrength += elem.strength
	}
	return totalFitness / totalStrength
}

// Modify the image in-place by swapping pixels to places where they match their neighbors.
func (i *MyImage) Congregate(kernelRadius int, numIters float64) {
	kernel := makeKernel(kernelRadius)
	numPixels := int(numIters * float64(i.xres*i.yres))

	for ii := 0; ii < numPixels; ii++ {
		if ii%2000 == 0 {
			pctDone := float64(int(float64(ii)/float64(numPixels)*1000)) / 10
			fmt.Println(pctDone)
		}

		// choose two random pixels
		x1 := RNG.Intn(i.xres)
		y1 := RNG.Intn(i.yres)
		x2 := RNG.Intn(i.xres)
		y2 := RNG.Intn(i.yres)
		if x1 == x2 && y1 == y2 {
			ii -= 1
			continue
		}
		c1 := i.pixels[x1][y1]
		c2 := i.pixels[x2][y2]

		// if swapping them would improve their total fitness, swap them
		originalFitness := i.pixelFitness(c1, x1, y1, kernel) + i.pixelFitness(c2, x2, y2, kernel)
		swappedFitness := i.pixelFitness(c2, x1, y1, kernel) + i.pixelFitness(c1, x2, y2, kernel)
		if swappedFitness > originalFitness {
			i.pixels[x1][y1] = c2
			i.pixels[x2][y2] = c1
		}
	}
}
