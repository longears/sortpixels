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
func sortPixelsValue(inFn, outFn string) {
	myImage := myimage.MakeMyImageFromPath(inFn)

	fmt.Println("  sorting using value")
	for ii := 0; ii < N_SORTS; ii++ {
		myImage.SortColumns("v", THREADPOOL_SIZE)
		myImage.SortRows("h2", THREADPOOL_SIZE)
	}
	myImage.SortColumns("v", THREADPOOL_SIZE)

	myImage.SaveAs(outFn)
}

func sortPixelsSaturationValue(inFn, outFn string) {
	myImage := myimage.MakeMyImageFromPath(inFn)

	fmt.Println("  sorting using saturation and value")
	for ii := 0; ii < N_SORTS; ii++ {
		myImage.SortColumns("sv", THREADPOOL_SIZE)
		myImage.SortRows("h2", THREADPOOL_SIZE)
	}
	myImage.SortColumns("sv", THREADPOOL_SIZE)

	myImage.SaveAs(outFn)
}

func congregatePixels(inFn, outFn string) {
	myImage := myimage.MakeMyImageFromPath(inFn)

	fmt.Println("  resizing")
	myImage = myImage.ThumbnailByPixels(512)

	fmt.Println("  scrambling")
	myImage.SortColumns("random", THREADPOOL_SIZE)
	myImage.SortRows("random", THREADPOOL_SIZE)

	fmt.Println("  congregating (large scale)")
	myImage.Congregate(0, 55) // maxMoveDist, percent of image visited per iteration
	fmt.Println("  congregating (small scale)")
	myImage.Congregate(8, 75) // maxMoveDist, percent of image visited per iteration

	myImage.SaveAs(outFn)
}

func transformImage(inFn, tag string, transformFunction func(string, string)) {
	// build outFn from inFn
	outFn := inFn
	if strings.Contains(outFn, ".") {
		dotii := strings.LastIndex(outFn, ".")
		outFn = outFn[:dotii] + "." + tag + ".png"
	} else {
		outFn += "." + tag
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
		transformFunction(inFn, outFn)
	}

	// attempt to give memory back to the OS
	debug.FreeOSMemory()

	fmt.Println()
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

		transformImage(inFn, "sorted_v", sortPixelsValue)
		transformImage(inFn, "sorted_sv", sortPixelsSaturationValue)
		transformImage(inFn, "congregated", congregatePixels)
	}
}
