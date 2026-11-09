package main

import (
	"fmt"
	"image/color"
	"os"
	"testing"
)

func TestInspectCaptures(t *testing.T) {
	err := os.Chdir("../..")
	if err != nil {
		t.Fatal(err)
	}

	active, err := readCapture("09_frames")
	if err != nil {
		t.Fatal(err)
	}

	inactive, err := readCapture("00_inactive")
	if err != nil {
		t.Fatal(err)
	}

	caption, err := readCapture("08_caption")
	if err != nil {
		t.Fatal(err)
	}

	colors := make(map[color.Color]color.Color)
	conflicts := 0
	for _, left := range []int{923, 946} {
		for vertical := range 21 {
			for horizontal := range 21 {
				first := active.At(left+horizontal, 92+vertical)
				second := inactive.At(left+horizontal, 92+vertical)
				if previous, ok := colors[first]; ok && previous != second {
					conflicts++
				}

				colors[first] = second
			}
		}
	}

	missing := make(map[color.Color]int)
	for vertical := range 21 {
		for horizontal := range 21 {
			pixel := caption.At(308+horizontal, 512+vertical)
			if _, ok := colors[pixel]; !ok {
				missing[pixel]++
			}
		}
	}

	fmt.Println("inactive palette", len(colors), "conflicts", conflicts, "missing restore", missing)
	for vertical := range 21 {
		for horizontal := range 21 {
			if active.At(923+horizontal, 92+vertical) != caption.At(308+horizontal, 296+vertical) {
				fmt.Println("caption background difference", horizontal, vertical, active.At(923+horizontal, 92+vertical), caption.At(308+horizontal, 296+vertical))
			}
		}
	}
}
