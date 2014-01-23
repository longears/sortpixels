package utils

import (
	"os"
	"sort"
)

// Check if a path exists or not.
func PathExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		} else {
			panic(err)
		}
	}
	return true
}

// return val clamped to be between min and max (inclusive)
func IntClamp(val int, min int, max int) int {
	if val < min {
		return min
	} else if val > max {
		return max
	}
	return val
}

func IntMax(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func IntMin(a int, b int) int {
	if a < b {
		return a
	}
	return b
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
