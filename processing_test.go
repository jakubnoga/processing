package processing

import (
	"bufio"
	"image"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	_ "image/jpeg"

	_ "image/png"

	_ "image/gif"
)

func TestPalette_toSlice(t *testing.T) {
	tests := []struct {
		name    string
		palette color.Palette
		want    [][]uint32
	}{
		{
			name:    "Basic test",
			palette: color.Palette{color.NRGBA{10, 20, 30, 255}, color.NRGBA{30, 20, 30, 255}},
			want:    [][]uint32{{10, 20, 30, 255}, {30, 20, 30, 255}},
		},
		{
			name:    "Basic test",
			palette: color.Palette{color.NRGBA{10, 20, 30, 255}, color.NRGBA{30, 20, 30, 255}},
			want:    [][]uint32{{10, 20, 30, 255}, {30, 20, 30, 255}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := paletteToSlice(tt.palette); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Palette.toSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func convert(imageName string, paletteName string) (have image.Image, want image.Image) {

	path, _ := filepath.Abs(imageName)
	file, e := os.Open(path)
	if e != nil {
		log.Fatalf("%v", e)
	}
	defer file.Close()

	palettePath, _ := filepath.Abs(paletteName)
	palettefile, e := os.Open(palettePath)
	if e != nil {
		log.Fatalf("%v", e)
	}
	defer palettefile.Close()

	img, _, e := image.Decode(bufio.NewReader(file))
	if e != nil {
		log.Fatalf("%v", e)
	}

	hexReader := new(HexReader)
	palette, _ := hexReader.Read(palettefile)

	return NewKdTreeProcessor(palette).ConvertImage(img), NewNaiveProcessor(palette).ConvertImage(img)
}

func TestHexReader_Read(t *testing.T) {
	path, _ := filepath.Abs("./assets/commodore64.hex")
	file, e := os.Open(path)
	if e != nil {
		log.Fatalf("%v", e)
	}
	defer file.Close()

	type args struct {
		file *os.File
	}
	tests := []struct {
		name    string
		hr      *HexReader
		args    args
		want    color.Palette
		wantErr bool
	}{
		{
			"Commodore64.hex",
			new(HexReader),
			args{file},
			color.Palette{
				color.NRGBA{0, 0, 0, 255},
				color.NRGBA{98, 98, 98, 255},
				color.NRGBA{137, 137, 137, 255},
				color.NRGBA{173, 173, 173, 255},
				color.NRGBA{255, 255, 255, 255},
				color.NRGBA{159, 78, 68, 255},
				color.NRGBA{203, 126, 117, 255},
				color.NRGBA{109, 84, 18, 255},
				color.NRGBA{161, 104, 60, 255},
				color.NRGBA{201, 212, 135, 255},
				color.NRGBA{154, 226, 155, 255},
				color.NRGBA{92, 171, 94, 255},
				color.NRGBA{106, 191, 198, 255},
				color.NRGBA{136, 126, 203, 255},
				color.NRGBA{80, 69, 155, 255},
				color.NRGBA{160, 87, 163, 255},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.hr.Read(tt.args.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("HexReader.Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HexReader.Read() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKdTreeProcessor_ConvertImage(t *testing.T) {
	type test struct {
		name    string
		input   string
		palette string
	}

	tests := []test{}

	images, _ := filepath.Glob("./assets/*.jpg")
	palettes, _ := filepath.Glob("./assets/*.hex")

	for _, image := range images {
		for _, palette := range palettes {
			imageName := filepath.Base(image[:len(image)-4])
			paletteName := filepath.Base(palette[:len(palette)-4])

			tests = append(tests, test{imageName + "_" + paletteName, image, palette})
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, want := convert(tt.input, tt.palette)

			if !reflect.DeepEqual(got.Bounds(), want.Bounds()) {
				log.Fatalf("got.Bounds() == %v, want.Bounds() == %v", got.Bounds(), want.Bounds())
			}

			threshold := 1.0
			errors := 0.0
			numberOfPixels := float64(got.Bounds().Dx() * got.Bounds().Dy())

			for x := got.Bounds().Min.X; x <= got.Bounds().Max.X; x++ {
				for y := got.Bounds().Min.Y; y <= got.Bounds().Max.Y; y++ {
					if !reflect.DeepEqual(got.At(x, y), want.At(x, y)) {
						errors++
					}
				}
			}

			if errors/numberOfPixels > threshold {
				log.Fatalf("Number of errors (%v) exceeded threshold (%v)", errors/numberOfPixels, threshold)
			}

		})
	}
}
