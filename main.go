package main

import (
	"flag"
	"fmt"
	"image/gif"
	"io"
	"math"
	"os"
	"sync"
)

var isCircular bool
var numLeds int
var ledOffset int

func convertGIF(inputFilename string) {
	var g *gif.GIF
	var input *os.File
	var output *os.File
	var err error

	if input, err = os.Open(inputFilename); err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to open input file: %s\n", inputFilename)
		return
	}
	defer input.Close()

	if g, err = gif.DecodeAll(input); err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to decode GIF: %s\n", inputFilename)
		return
	}

	outputFilename := inputFilename + ".bin"
	if output, err = os.OpenFile(outputFilename, os.O_CREATE | os.O_WRONLY | os.O_TRUNC, 0600); err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to open output file: %s\n", outputFilename)
		return
	}
	defer output.Close()

	if isCircular {
		convertGIFCircular(output, g)
	} else {
		convertGIFRectangular(output, g)
	}
}

func convertGIFRectangular(output io.Writer, g *gif.GIF) {
	for _, image := range g.Image {
		for y := 0; y < image.Rect.Max.Y; y++ {
			for x := 0; x < image.Rect.Max.X; x++ {
				r, g, b, a := image.At(x, y).RGBA()
				r = r * a / 255
				g = g * a / 255
				b = b * a / 255
				output.Write([]byte{
					byte(r),
					byte(g),
					byte(b),
				})
			}
		}
	}
	
}

func convertGIFCircular(output io.Writer, g *gif.GIF) {
	for _, image := range g.Image {
		centerX := float64((image.Rect.Max.X - image.Rect.Min.X) / 2)
		centerY := float64((image.Rect.Max.Y - image.Rect.Min.Y) / 2)

		radius := centerX
		if centerX > centerY {
			radius = centerY
		}

		for i := 0; i < 360; i++ {
			dx := radius * math.Cos(float64(i) / 180 * math.Pi) / float64(numLeds)
			dy := radius * math.Sin(float64(i) / 180 * math.Pi) / float64(numLeds)
			offsetX := dx * float64(ledOffset)
			offsetY := dy * float64(ledOffset)
			offsetRatio := float64(numLeds - ledOffset) / float64(numLeds)
			dx *= offsetRatio
			dy *= offsetRatio
			x := centerX + offsetX
			y := centerY + offsetY

			for j := 0; j < numLeds; j++ {
				r, g, b, a := image.At(int(x), int(y)).RGBA()
				r = r * a / 255
				g = g * a / 255
				b = b * a / 255
				output.Write([]byte{
					byte(r),
					byte(g),
					byte(b),
				})

				x += dx
				y += dy
			}
		}
	}
	
}

func init() {
	flag.BoolVar(&isCircular, "circular", false, "pack pixels in higher-res and circular way")
	flag.IntVar(&numLeds, "num-leds", 0, "set number of leds (only required when using -circular)")
	flag.IntVar(&ledOffset, "led-offset", 0, "set led offset (only used when using -circular)")
}

func main() {
	var wg sync.WaitGroup

	flag.Parse()

	if isCircular && numLeds == 0 {
		fmt.Println("must specify -numleds higher than 0 when using -circular")
		return
	}

	for _, filename := range flag.Args() {
		wg.Add(1)

		go func() {
			convertGIF(filename)
			wg.Done()
		}()
	}

	wg.Wait()
}
