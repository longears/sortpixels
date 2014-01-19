package main

import (
	"fmt"
	"github.com/longears/sortpixels/myimage"
	"image"
	"image/png"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
)

// How many times to repeat the vertical & horizontal sort step
const N_SORTS = 6

// How many threads to run in parallel
var THREADPOOL_SIZE int

func init() {
	THREADPOOL_SIZE = runtime.NumCPU()
	runtime.GOMAXPROCS(runtime.NumCPU())
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
	myImage := &myimage.MyImage{}
	myImage.PopulateFromImage(img)
	img = nil

	// sort
	fmt.Println("  sorting")
	//myImage.SortRows("semirandom", THREADPOOL_SIZE)
	for ii := 0; ii < N_SORTS; ii++ {
		//fmt.Println("   ", ii+1, "/", N_SORTS)
		myImage.SortColumns("v", THREADPOOL_SIZE)
		myImage.SortRows("h2", THREADPOOL_SIZE)
	}
	myImage.SortColumns("v", THREADPOOL_SIZE)

	// convert back to built in image
	fmt.Println("  converting to built in image")
	destImg := myImage.ToBuiltInImage()
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
