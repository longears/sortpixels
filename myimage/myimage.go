package myimage

import (
	"fmt"
	"github.com/longears/sortpixels/mycolor"
	"github.com/longears/sortpixels/utils"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
)

//================================================================================
// IMAGE

type MyImage struct {
	xres   int
	yres   int
	pixels [][]*mycolor.MyColor // 2d array, [x][y]
}

// Init the MyImage pixel array, creating MyColor objects
// from the data in the given image (from the built-in image package).
// HSV is computed here also for each pixel.
func (i *MyImage) PopulateFromImage(img image.Image) {
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

func (i *MyImage) String() string {
	return fmt.Sprintf("<image %v x %v>", i.xres, i.yres)
}

// Read y coordinates over yChan and sort those rows.
// Send 1 to doneChan when each row is done.
// The image natively stores pixels in columns, not rows, so we
// have to copy the pixels into a temporary slice, sort it, then
// put it back.
func goSortRow(i *MyImage, kind string, yChan chan int, doneChan chan int) {
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
		go goSortRow(i, kind, yChan, doneChan)
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
