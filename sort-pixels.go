package main

import (
	"fmt"
	"github.com/longears/sortpixels/myimage"
	"github.com/longears/sortpixels/utils"
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
// IMAGE MODIFICATION ALGORITHMS

// Read the image from the path inFn,
// sort the pixels,
// and save the result to the path outFn.
// Return an error if the input file is not decodable as an image.
func sortPixels(inFn, outFn string) {
	myImage := myimage.MakeMyImageFromPath(inFn)

	fmt.Println("  sorting")
	for ii := 0; ii < N_SORTS; ii++ {
		myImage.SortColumns("v", THREADPOOL_SIZE)
		myImage.SortRows("h2", THREADPOOL_SIZE)
	}
	myImage.SortColumns("v", THREADPOOL_SIZE)

	myImage.SaveAs(outFn)
}

func congregatePixels(inFn, outFn string) {
	myImage := myimage.MakeMyImageFromPath(inFn)
	myImage = myImage.Thumbnail(0.2)

	fmt.Println("  scrambling")
	myImage.SortColumns("random", THREADPOOL_SIZE)
	myImage.SortRows("random", THREADPOOL_SIZE)

	fmt.Println("  congregating")
	for ii := 0; ii < 1; ii++ {
		myImage.Congregate(50) // thumb size in pixels, percent of image visited per iteration
		tempFn := outFn + "." + fmt.Sprintf("%03d", ii) + ".png"
		fmt.Println(tempFn)
		myImage.SaveAs(tempFn)
	}

	//fmt.Println("  showing thumb")
	//myImage.ShowThumb(0.015)

	myImage.SaveAs(outFn)
}

//================================================================================
// MAIN

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
	if !utils.PathExists("output") {
		err := os.Mkdir("output", 0755)
		if err != nil {
			panic(fmt.Sprintf("%v", err))
		}
	}

	// open, sort, and save input images
	for inputII := 1; inputII < len(os.Args); inputII++ {
		inFn := os.Args[inputII]

		// build outFn from inFn
		outFn := inFn
		if strings.Contains(outFn, ".") {
			dotii := strings.LastIndex(outFn, ".")
			outFn = outFn[:dotii] + ".sorted.png"
		} else {
			outFn += ".sorted"
		}
		if strings.Contains(outFn, "/") {
			outFn = outFn[strings.LastIndex(outFn, "/")+1:]
		}
		outFn = "output/" + outFn

		// read, sort, and save (unless file has already been sorted)
		fmt.Println(inFn)
		if utils.PathExists(outFn) {
			fmt.Println("  SKIPPING: already exists")
		} else {
			//sortPixels(inFn, outFn)
			congregatePixels(inFn, outFn)
		}

		// attempt to give memory back to the OS
		debug.FreeOSMemory()

		fmt.Println()
	}
}
