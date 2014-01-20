sortpixels
==========

Sort and rearrange the pixels in an image.

![](http://birdhat.org/misc/sort-pixels/img/sorted.jpg)

The algorithm
---------------------

1. Take each column of an image and sort those pixels vertically by brightness.
2. Take each row and sort the pixels horizontally by hue.

Repeat these two steps until the image converges on a stable arrangement of pixels.  You'll get an abstract version of the original image using the exact same pixels, but clumped together into areas of similar color.

This was first implemented with in-browser Javascript as [The Image Pixel Sorter.](http://birdhat.org/misc/sort-pixels/)  This version is written in Go and is much faster.

There is also a [blog of sorted images.](http://sorted-pixels.tumblr.com/)

Usage
---------------------

`sort-pixels input1.png [input2.jpg input3.png ...]`

The resulting sorted images will be written into the `./output` subfolder (which is created if needed).  Images which are already in the output folder will be skipped.

**Warning:** very large images (e.g. 2000 x 4000 pixels) can use 500 MB of memory during sorting.

