package mycolor

import (
	"math/rand"
)

//================================================================================
// COLOR

type MyColor struct {
	R         uint8
	G         uint8
	B         uint8
	A         uint8
	H         float32
	S         float32
	V         float32
	SortValue float32
}

// Compute and set the SortValue for the MyColor object.
// "kind" is the type of sort to do.  Use one of: random semirandom h h2 v s
func (c *MyColor) SetSortValue(kind string, ii int) {
	switch kind {
	case "random":
		// totally randomize the order of the pixels
		c.SortValue = rand.Float32()
	case "semirandom":
		// move pixels plus or minus 100 pixels
		c.SortValue = float32(ii)/4 + rand.Float32()*25
	case "h":
		c.SortValue = c.H
	case "h2":
		// sort by hue unless saturation is too low.
		// unsaturated pixels will sort to the front.
		c.SortValue = c.H + 0.15
		if c.SortValue > 1 {
			c.SortValue -= 1
		}
		if c.S < 0.07 {
			c.SortValue -= 900
		}
	case "v":
		c.SortValue = -(float32(c.R)/255*0.30 + float32(c.G)/255*0.59 + float32(c.B)/255*0.11)
	case "s":
		c.SortValue = c.S
	default:
		panic("bad sort kind: " + kind)
	}
}

func threeMax(a float32, b float32, c float32) float32 {
	if a > b {
		if a > c {
			return a
		} else {
			return c
		}
	} else {
		if b > c {
			return b
		} else {
			return c
		}
	}
}

func threeMin(a float32, b float32, c float32) float32 {
	if a < b {
		if a < c {
			return a
		} else {
			return c
		}
	} else {
		if b < c {
			return b
		} else {
			return c
		}
	}
}

// Read r, g b in the range 0-255; set h, s, v in the range 0-1.
// Taken from http://stackoverflow.com/questions/8022885/rgb-to-hsv-color-in-javascript
func (c *MyColor) ComputeHSV() {
	var h, s, v float32

	r := float32(c.R) / 255
	g := float32(c.G) / 255
	b := float32(c.B) / 255

	v = threeMax(r, g, b)
	diff := v - threeMin(r, g, b)

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
	c.H = h
	c.S = s
	c.V = v
}
