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

// Return a new image which has a size of (ratio * original_image_size)
func (i *MyImage) Thumbnail(ratio float64) *MyImage {
	thumb := &MyImage{}
	thumb.xres = int(float64(i.xres) * ratio)
	thumb.yres = int(float64(i.yres) * ratio)
	thumb.pixels = make([][]*mycolor.MyColor, thumb.xres)

	// for each pixel in the thumbnail...
	for tx := 0; tx < thumb.xres; tx++ {
		thumb.pixels[tx] = make([]*mycolor.MyColor, thumb.yres)
		originalMinX := int(float64(tx) / float64(thumb.xres) * float64(i.xres))
		originalMaxX := int(float64(tx+1) / float64(thumb.xres) * float64(i.xres))
		for ty := 0; ty < thumb.yres; ty++ {
			originalMinY := int(float64(ty) / float64(thumb.yres) * float64(i.yres))
			originalMaxY := int(float64(ty+1) / float64(thumb.yres) * float64(i.yres))
			// average together the corresponding pixels in the original
			totalR := 0
			totalG := 0
			totalB := 0
			numPixels := 0
			for ox := originalMinX; ox < originalMaxX; ox++ {
				for oy := originalMinY; oy < originalMaxY; oy++ {
					c := i.pixels[ox][oy]
					totalR += int(c.R)
					totalG += int(c.G)
					totalB += int(c.B)
					numPixels += 1
				}
			}
			avgColor := &mycolor.MyColor{}
			avgColor.R = uint8(float64(totalR) / float64(numPixels))
			avgColor.G = uint8(float64(totalG) / float64(numPixels))
			avgColor.B = uint8(float64(totalB) / float64(numPixels))
			avgColor.A = 255 // TODO: handle alpha more carefully
			avgColor.ComputeHSV()
			thumb.pixels[tx][ty] = avgColor
		}
	}

	return thumb
}

// Return an interpolated pixel value from the image.
// x and y values outside the image will return the value from the
// nearest point in the image.
func (i *MyImage) GetColorWithLinearInterpolation(x float64, y float64) *mycolor.MyColor {
	// the 4 neighboring pixel indices
	x0 := utils.IntClamp(int(x-0.5), 0, i.xres-1)
	x1 := utils.IntClamp(int(x+0.5), 0, i.xres-1)
	y0 := utils.IntClamp(int(y-0.5), 0, i.yres-1)
	y1 := utils.IntClamp(int(y+0.5), 0, i.yres-1)

	// percent of the way from x0 to x1
	xPct := (x - 0.5) - float64(int(x-0.5))
	yPct := (y - 0.5) - float64(int(y-0.5))
	// get pixels and interpolate
	c00 := i.pixels[x0][y0]
	c01 := i.pixels[x0][y1]
	c10 := i.pixels[x1][y0]
	c11 := i.pixels[x1][y1]
	r := uint8((float64(c00.R)*(1-xPct)+float64(c10.R)*xPct)*(1-yPct) + (float64(c01.R)*(1-xPct)+float64(c11.R)*xPct)*yPct)
	g := uint8((float64(c00.G)*(1-xPct)+float64(c10.G)*xPct)*(1-yPct) + (float64(c01.G)*(1-xPct)+float64(c11.G)*xPct)*yPct)
	b := uint8((float64(c00.B)*(1-xPct)+float64(c10.B)*xPct)*(1-yPct) + (float64(c01.B)*(1-xPct)+float64(c11.B)*xPct)*yPct)
	cResult := &mycolor.MyColor{r, g, b, 255, 0, 0, 0, 0}
	cResult.ComputeHSV()
	return cResult
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

// Return a float between 0 and 1 indicating how similar the colors are.
func colorSimilarity(a *mycolor.MyColor, b *mycolor.MyColor) float64 {
	return 1.0 - (1.5*math.Abs(float64(a.H-b.H))+0.5*math.Abs(float64(a.S-b.S))+math.Abs(float64(a.V-b.V)))/3.0
}

func (i *MyImage) thumbPixelFitness(color *mycolor.MyColor, x int, y int, thumb *MyImage) float64 {
	thumbX := float64(x) * float64(thumb.xres) / float64(i.xres)
	thumbY := float64(y) * float64(thumb.yres) / float64(i.yres)
	thumbColor := thumb.GetColorWithLinearInterpolation(thumbX, thumbY)
	return colorSimilarity(color, thumbColor)
}

func (i *MyImage) colorPosPixelFitness(color *mycolor.MyColor, x int, y int, thumb *MyImage) float64 {
	idealRad := 1 - (float64(color.S)+float64(color.V))/2
	idealRad *= 1
	idealTheta := float64(color.H) * 2 * math.Pi
	idealX := math.Cos(idealTheta) * idealRad // 0 at the top
	idealY := math.Sin(idealTheta) * idealRad
	idealX = (idealX/2.0 + 0.5) * float64(i.xres)
	idealY = (idealY/2.0 + 0.5) * float64(i.yres)
	//idealX := float64(color.H) * float64(i.xres)
	//idealY := float64(color.V) * float64(i.yres)
	return -math.Pow(math.Hypot(float64(x)-idealX, float64(y)-idealY), 2)
}

// Modify the image in-place by swapping pixels to places where they match their neighbors.
func (i *MyImage) Congregate(thumbPixels int, numIters float64) {
	thumbRatio := (float64(thumbPixels) + 0.01) / float64(utils.IntMin(i.xres, i.yres))
	thumb := i.Thumbnail(thumbRatio)

	numPixels := int(numIters * float64(i.xres*i.yres))
	for ii := 0; ii < numPixels; ii++ {
		if ii%300000 == 0 {
			pctDone := float64(int(float64(ii)/float64(numPixels)*1000)) / 10
			fmt.Println(pctDone)
		}

		//// occasionally re-make the thumbnail
		//if ii%1000 == 0 {
		//    thumb := i.Thumbnail(0.1)
		//}

		// choose two random pixels
		x1 := rand.Intn(i.xres)
		y1 := rand.Intn(i.yres)
		x2 := rand.Intn(i.xres)
		y2 := rand.Intn(i.yres)
		if x1 == x2 && y1 == y2 {
			ii -= 1
			continue
		}
		c1 := i.pixels[x1][y1]
		c2 := i.pixels[x2][y2]

		// if swapping them would improve their total fitness, swap them
		originalFitness := i.colorPosPixelFitness(c1, x1, y1, thumb) + i.colorPosPixelFitness(c2, x2, y2, thumb)
		swappedFitness := i.colorPosPixelFitness(c2, x1, y1, thumb) + i.colorPosPixelFitness(c1, x2, y2, thumb)
		if swappedFitness > originalFitness {
			i.pixels[x1][y1] = c2
			i.pixels[x2][y2] = c1
		}
	}

}

func (i *MyImage) ShowThumb(thumbSize float64) {
	thumb := i.Thumbnail(thumbSize)

	for x := 0; x < i.xres; x++ {
		for y := 0; y < i.yres; y++ {
			thumbX := float64(x) * float64(thumb.xres) / float64(i.xres)
			thumbY := float64(y) * float64(thumb.yres) / float64(i.yres)
			thumbColor := thumb.GetColorWithLinearInterpolation(thumbX, thumbY)
			i.pixels[x][y] = thumbColor
		}
	}
}
