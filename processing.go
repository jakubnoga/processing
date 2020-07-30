package processing

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"github.com/jakubnoga/kdtree"
	"os"
	"strconv"
)

// Processor provides image processing methods
type Processor interface {
	Convert(color color.Color) color.Color
	ConvertImage(image image.Image) image.Image
}

// ConvertImage using provided processor
func ConvertImage(input image.Image, p Processor) image.Image {
	var output = image.NewNRGBA(input.Bounds())

	for x := input.Bounds().Min.X; x <= input.Bounds().Max.X; x++ {
		for y := input.Bounds().Min.Y; y <= input.Bounds().Max.Y; y++ {
			output.Set(x, y, p.Convert(input.At(x, y)))
		}
	}

	return output
}

type kdTreeProcessor struct {
	tree *kdtree.KdTree
}

// NewKdTreeProcessor initialized with palette
func NewKdTreeProcessor(palette color.Palette) Processor {
	return &kdTreeProcessor{kdtree.Create(paletteToSlice(palette), 0)}
}

func (p *kdTreeProcessor) Convert(color color.Color) color.Color {
	return arrayToColor(p.tree.NearestNeighbour(colorToArray(color)).Point)
}

func (p *kdTreeProcessor) ConvertImage(input image.Image) image.Image {
	return ConvertImage(input, p)
}

type naiveProcessor struct {
	palette color.Palette
}

// NewNaiveProcessor initialized with palette
func NewNaiveProcessor(palette color.Palette) Processor {
	return &naiveProcessor{palette}
}

func (p *naiveProcessor) Convert(color color.Color) color.Color {
	return p.palette.Convert(color)
}

func (p *naiveProcessor) ConvertImage(input image.Image) image.Image {
	return ConvertImage(input, p)
}

// PaletteReader reads palette from file
type PaletteReader interface {
	Read(file *os.File) color.Palette
}

// HexReader reads palette from file containing RGB values in hex one color per line
type HexReader struct{}

func (hr *HexReader) Read(file *os.File) (color.Palette, error) {
	scanner := bufio.NewScanner(bufio.NewReader(file))

	scanResult := scanner.Scan()
	palette := make([]color.Color, 0)

	for scanResult {
		line := scanner.Text()
		if len(line) != 6 {
			return nil, fmt.Errorf("Expected 6 chars per line, got %s", line)
		}

		var rgb [3]uint8
		for i := 0; i < 6; i = i + 2 {
			j := i / 2
			ui64, err := strconv.ParseUint(line[i:i+2], 16, 8)

			if err != nil {
				return nil, err
			}

			rgb[j] = uint8(ui64)
		}

		color := color.NRGBA{rgb[0], rgb[1], rgb[2], 255}
		palette = append(palette, color)

		scanResult = scanner.Scan()
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("Error while reading file %s: %w", file.Name(), err)
	}

	return palette, nil
}

func paletteToSlice(palette color.Palette) [][]uint32 {
	slice := make([][]uint32, len(palette))
	for idx, color := range palette {
		slice[idx] = colorToArray(color)
	}

	return slice
}

func colorToArray(color color.Color) []uint32 {
	r, g, b, a := color.RGBA()
	return []uint32{
		uint32(r >> 8),
		uint32(g >> 8),
		uint32(b >> 8),
		uint32(a >> 8),
	}
}

func arrayToColor(array []uint32) color.Color {
	return color.NRGBA{uint8(array[0]), uint8(array[1]), uint8(array[2]), uint8(array[3])}
}
