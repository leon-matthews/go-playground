// Command dhash prints the difference hash of an image, or compares two images.
package main

import (
	"flag"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"local.dev/dhash-go"
)

func main() {
	format := flag.String("format", "hex", "output format: hex, decimal, matrix, or grays")
	flag.Parse()

	var err error
	switch files := flag.Args(); len(files) {
	case 1:
		err = single(files[0], *format)
	case 2:
		err = compare(files[0], files[1])
	default:
		fmt.Fprintln(os.Stderr, "usage: dhash [-format hex|decimal|matrix|grays] FILE [FILE]")
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "dhash:", err)
		os.Exit(1)
	}
}

// single hashes one image and prints it in the requested format.
func single(name, format string) error {
	img, err := loadImage(name)
	if err != nil {
		return err
	}
	switch format {
	case "hex":
		fmt.Println(dhash.New(img))
	case "decimal":
		h := dhash.New(img)
		fmt.Println(h.Row, h.Col)
	case "matrix":
		h := dhash.New(img)
		fmt.Println(dhash.Matrix(h.Row, ". ", "* "))
		fmt.Println()
		fmt.Println(dhash.Matrix(h.Col, ". ", "* "))
	case "grays":
		fmt.Println(dhash.FormatGrays(dhash.Grays(img)))
	default:
		return fmt.Errorf("unknown format %q", format)
	}
	return nil
}

// compare hashes two images and reports how many bits differ.
func compare(name1, name2 string) error {
	img1, err := loadImage(name1)
	if err != nil {
		return err
	}
	img2, err := loadImage(name2)
	if err != nil {
		return err
	}

	different := dhash.New(img1).Distance(dhash.New(img2))
	unit := "bits differ"
	if different == 1 {
		unit = "bit differs"
	}
	fmt.Printf("%d %s out of 128 (%.1f%%)\n", different, unit, 100*float64(different)/128)
	return nil
}

// loadImage opens and decodes an image file.
func loadImage(name string) (image.Image, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	return img, err
}
