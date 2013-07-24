package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"math"
	"math/rand"
	"os"
	"runtime"
	"runtime/debug"
	"sort"
	"strings"
)

// How many times to repeat the vertical & horizontal sort step
const N_SORTS = 6

// How many threads to run in parallel
var THREADPOOL_SIZE int

// The random number generator
var RNG *rand.Rand

func init() {
	THREADPOOL_SIZE = runtime.NumCPU()
	runtime.GOMAXPROCS(runtime.NumCPU())
	RNG = rand.New(rand.NewSource(99))
}

//================================================================================
// SORTING RIGAMAROLE
// This is taken from https://github.com/daviddengcn/go-villa/blob/master/sort.go

type sortI struct {
	l    int
	less func(int, int) bool
	swap func(int, int)
}

func (s *sortI) Len() int {
	return s.l
}

func (s *sortI) Less(i, j int) bool {
	return s.less(i, j)
}

func (s *sortI) Swap(i, j int) {
	s.swap(i, j)
}

// SortF sorts the data defined by the length, Less and Swap functions.
func SortF(Len int, Less func(int, int) bool, Swap func(int, int)) {
	sort.Sort(&sortI{l: Len, less: Less, swap: Swap})
}

//================================================================================
// COLOR

type MyColor struct {
	r         uint8
	g         uint8
	b         uint8
	a         uint8
	h         float64
	s         float64
	v         float64
	sortValue float64
}

// Compute and set the sortValue for the MyColor object.
// "kind" is the type of sort to do.  Use one of: random semirandom h h2 v s
func (c *MyColor) setSortValue(kind string, ii int) {
	switch kind {
	case "random":
		// totally randomize the order of the pixels
		c.sortValue = RNG.Float64()
	case "semirandom":
		// move pixels plus or minus 100 pixels
		c.sortValue = float64(ii)/4 + RNG.Float64()*25
	case "h":
		c.sortValue = c.h
	case "h2":
		// sort by hue unless saturation is too low.
		// unsaturated pixels will sort to the front.
		c.sortValue = c.h + 0.15
		if c.sortValue > 1 {
			c.sortValue -= 1
		}
		if c.s < 0.07 {
			c.sortValue -= 900
		}
	case "v":
		c.sortValue = -(float64(c.r)/255*0.30 + float64(c.g)/255*0.59 + float64(c.b)/255*0.11)
	case "s":
		c.sortValue = c.s
	default:
		panic("bad sort kind: " + kind)
	}
}

// Read r, g b in the range 0-255; set h, s, v in the range 0-1.
// Taken from http://stackoverflow.com/questions/8022885/rgb-to-hsv-color-in-javascript
func (c *MyColor) computeHSV() {
	var h, s, v float64

	r := float64(c.r) / 255
	g := float64(c.g) / 255
	b := float64(c.b) / 255

	v = math.Max(r, math.Max(g, b))
	diff := v - math.Min(r, math.Min(g, b))

	if diff == 0 {
		h = 0
		s = 0
	} else {
		s = diff / v
		rr := (v-r)/6.0/diff + 0.5
		gg := (v-g)/6.0/diff + 0.5
		bb := (v-b)/6.0/diff + 0.5
		if r == v {
			h = bb - gg
		} else if g == v {
			h = 1.0/3.0 + rr - bb
		} else if b == v {
			h = 2.0/3.0 + gg - rr
		}

		if h < 0 {
			h += 1
		} else if h > 1 {
			h -= 1
		}
	}
	c.h = h
	c.s = s
	c.v = v
}

//================================================================================
// IMAGE

type MyImage struct {
	xres   int
	yres   int
	pixels [][]*MyColor // 2d array, [x][y]
}

// Init the MyImage pixel array, creating MyColor objects
// from the data in the given image (from the built-in image package).
// HSV is computed here also for each pixel.
func (i *MyImage) populateFromImage(img image.Image) {
	i.xres = img.Bounds().Max.X
	i.yres = img.Bounds().Max.Y
	i.pixels = make([][]*MyColor, i.xres)
	for x := 0; x < i.xres; x++ {
		i.pixels[x] = make([]*MyColor, i.yres)
		for y := 0; y < i.yres; y++ {
			r, g, b, a := img.At(x, y).RGBA()
			c := &MyColor{uint8(r / 256), uint8(g / 256), uint8(b / 256), uint8(a / 256), 0, 0, 0, 0}
			c.computeHSV()
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
	row := make([]*MyColor, i.xres)
	for y := range yChan {
		// copy into temp slice
		// set sort value
		for x := 0; x < i.xres; x++ {
			row[x] = i.pixels[x][y]
			row[x].setSortValue(kind, x)
		}
		// sort
		SortF(
			len(row),
			func(a, b int) bool {
				return row[a].sortValue < row[b].sortValue
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
func (i *MyImage) SortRows(kind string) {
	yChan := make(chan int, i.yres+10)
	doneChan := make(chan int, i.yres+10)
	for threadNum := 0; threadNum < THREADPOOL_SIZE; threadNum++ {
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
func goSortMyColorSlice(kind string, toSortChan chan []*MyColor, doneChan chan int) {
	for colorSlice := range toSortChan {
		// set sort value
		for ii, v := range colorSlice {
			v.setSortValue(kind, ii)
		}
		// do actual sort
		SortF(
			len(colorSlice),
			func(a, b int) bool {
				return colorSlice[a].sortValue < colorSlice[b].sortValue
			},
			func(a, b int) {
				colorSlice[a], colorSlice[b] = colorSlice[b], colorSlice[a]
			})
		doneChan <- 1
	}
}

// Launch some threads to sort the rows of the image.
// Wait until complete, and kill the threads.
func (i *MyImage) SortColumns(kind string) {
	toSortChan := make(chan []*MyColor, i.xres+10)
	doneChan := make(chan int, i.xres+10)
	for threadNum := 0; threadNum < THREADPOOL_SIZE; threadNum++ {
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
func (i *MyImage) toBuiltInImage() image.Image {
	destImg := image.NewRGBA(image.Rectangle{image.ZP, image.Point{i.xres, i.yres}})
	for x := 0; x < i.xres; x++ {
		for y := 0; y < i.yres; y++ {
			myColor := i.pixels[x][y]
			rgba := color.RGBA{uint8(myColor.r), uint8(myColor.g), uint8(myColor.b), uint8(myColor.a)}
			destImg.Set(x, y, rgba)
		}
	}
	return destImg
}

//================================================================================
// MAIN

func handleErr(err error) {
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}
}

// Read the image from the path inFn,
// sort the pixels,
// and save the result to the path outFn.
// Return an error if the input file is not decodable as an image.
func sortPixels(inFn, outFn string) error {
	// open file and decode image
	fmt.Println("  reading and decoding image")
	file, err := os.Open(inFn)
	handleErr(err)
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		// couldn't decode; probably the file is not actually an image
		return err
	}

	// convert to MyImage
	fmt.Println("  converting to MyImage")
	myImage := &MyImage{}
	myImage.populateFromImage(img)
	img = nil

	// sort
	fmt.Println("  sorting")
	//myImage.SortRows("semirandom")
	for ii := 0; ii < N_SORTS; ii++ {
		//fmt.Println("   ", ii+1, "/", N_SORTS)
		myImage.SortColumns("v")
		myImage.SortRows("h2")
	}
	myImage.SortColumns("v")

	// convert back to built in image
	fmt.Println("  converting to built in image")
	destImg := myImage.toBuiltInImage()
	myImage = nil

	// write output
	fmt.Println("  writing to", outFn)
	fo, err := os.Create(outFn)
	handleErr(err)
	defer func() {
		err := fo.Close()
		handleErr(err)
	}()
	png.Encode(fo, destImg)

	return nil
}

// Check if a path exists or not.
func Exists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		} else {
			panic(err)
		}
	}
	return true
}

func main() {
	fmt.Println("------------------------------------------------------------\\")
	defer fmt.Println("------------------------------------------------------------/")

	// handle command line
	if len(os.Args) < 2 {
		fmt.Println()
		fmt.Println("  usage:  sort  input.png  [input2.jpg input3.png ...]")
		fmt.Println()
		fmt.Println("  Sort the pixels in the image(s) and save to the ./output/ folder.")
		fmt.Println()
		return
	}

	// make output directory if needed
	if !Exists("output") {
		err := os.Mkdir("output", 0755)
		handleErr(err)
	}

	// open, sort, and save input images
	for inputII := 1; inputII < len(os.Args); inputII++ {
		inFn := os.Args[inputII]

		// build outFn from inFn
		outFn := inFn
		if strings.Contains(outFn, ".") {
			dotii := strings.LastIndex(outFn, ".")
			outFn = outFn[:dotii] + ".sorted." + outFn[dotii+1:]
		} else {
			outFn += ".sorted"
		}
		if strings.Contains(outFn, "/") {
			outFn = outFn[strings.LastIndex(outFn, "/")+1:]
		}
		outFn = "output/" + outFn

		// read, sort, and save (unless file has already been sorted)
		fmt.Println(inFn)
		if Exists(outFn) {
			fmt.Println("  SKIPPING: already exists")
		} else {
			err := sortPixels(inFn, outFn)
			if err != nil {
				fmt.Println("  oops, that wasn't an image.")
			}
		}

		// attempt to give memory back to the OS
		debug.FreeOSMemory()

		fmt.Println()
	}
}
