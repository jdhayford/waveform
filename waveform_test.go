package waveform

import (
	"bytes"
	"fmt"
	"image/color"
	"image/png"
	"io/ioutil"
	"log"
	"math"
	"os"
	"testing"

	"azul3d.org/audio.v1"
)

var (
	// Read in test files
	wavFile = func() []byte {
		file, err := ioutil.ReadFile("./test/tone16bit.wav")
		if err != nil {
			log.Fatalf("could not open test WAV: %v", err)
		}

		return file
	}()
	flacFile = func() []byte {
		file, err := ioutil.ReadFile("./test/tone16bit.flac")
		if err != nil {
			log.Fatalf("could not open test FLAC: %v", err)
		}

		return file
	}()
	mp3File = func() []byte {
		file, err := ioutil.ReadFile("./test/tone16bit.mp3")
		if err != nil {
			log.Fatalf("could not open test MP3: %v", err)
		}

		return file
	}()
	oggVorbisFile = func() []byte {
		file, err := ioutil.ReadFile("./test/tone16bit.ogg")
		if err != nil {
			log.Fatalf("could not open test Ogg Vorbis: %v", err)
		}

		return file
	}()
)

// ExampleNew provides example usage of waveform, using a media file from the filesystem.
func ExampleNew() {
	// waveform accepts io.Reader, so we will use a media file in the filesystem
	file, err := os.Open("./test/tone16bit.flac")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("open:", file.Name())
	defer file.Close()

	// Generate waveform image from audio file
	img, err := New(file, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Encode image as PNG into buffer
	buf := bytes.NewBuffer(nil)
	if err := png.Encode(buf, img); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("encoded: %d bytes", buf.Len())

	// Output:
	// open: ./test/tone16bit.flac
	// encoded: 86 bytes
}

// TestNew verifies that New creates the proper parser for an example input stream
func TestNew(t *testing.T) {
	// X is set by file duration, Y by library
	const defaultX = 4
	const defaultY = 128

	// Table of tests
	var tests = []struct {
		stream  []byte
		err     error
		options *Options
	}{
		// MP3 file
		{mp3File, ErrFormat, nil},
		// Ogg Vorbis file
		{oggVorbisFile, ErrFormat, nil},
		// Unknown format
		{[]byte("nonsense"), ErrFormat, nil},
		// WAV file, no options
		{wavFile, nil, nil},
		// WAV file, scaled
		{wavFile, nil, &Options{ScaleX: 5, ScaleY: 2}},
		// FLAC file, no options
		{flacFile, nil, nil},
		// FLAC file, scaled
		{flacFile, nil, &Options{ScaleX: 5, ScaleY: 2}},
	}

	// Iterate all tests
	for _, test := range tests {
		// Generate a io.Reader
		reader := bytes.NewReader(test.stream)

		// Attempt to create image for the reader
		img, err := New(reader, test.options)
		if err != nil {
			if err == test.err {
				continue
			}

			t.Fatalf("unexpected error: %v", err)
		}

		// Verify that image is RGBA
		if model := img.ColorModel(); model != color.RGBAModel {
			t.Fatalf("unexpected color model: %v != %v", model, color.RGBAModel)
		}

		// Check for expected bounds
		bounds := img.Bounds()
		var scaleX, scaleY int
		if test.options == nil {
			scaleX = defaultX
			scaleY = defaultY
		} else {
			// Set scale by options
			scaleX = test.options.ScaleX * defaultX
			scaleY = test.options.ScaleY * defaultY
		}

		if bounds.Max.X != scaleX || bounds.Max.Y != scaleY {
			t.Fatalf("unexpected bounds: (%v,%v) != (%v,%v)", bounds.Max.X, bounds.Max.Y, scaleX, scaleY)
		}
	}
}

// Test_rmsF64Samples verifies that rmsF64Samples computes correct results
func Test_rmsF64Samples(t *testing.T) {
	var tests = []struct {
		samples audio.F64Samples
		result  float64
		isNaN   bool
	}{
		// Empty samples - NaN
		{audio.F64Samples{}, 0.00, true},
		// Negative samples
		{audio.F64Samples{-0.10}, 0.10, false},
		{audio.F64Samples{-0.10, -0.20}, 0.15811388300841897, false},
		{audio.F64Samples{-0.10, -0.20, -0.30, -0.40, -0.50}, 0.33166247903554, false},
		// Positive samples
		{audio.F64Samples{0.10}, 0.10, false},
		{audio.F64Samples{0.10, 0.20}, 0.15811388300841897, false},
		{audio.F64Samples{0.10, 0.20, 0.30, 0.40, 0.50}, 0.33166247903554, false},
		// Mixed samples
		{audio.F64Samples{0.10}, 0.10, false},
		{audio.F64Samples{0.10, -0.20}, 0.15811388300841897, false},
		{audio.F64Samples{0.10, -0.20, 0.30, -0.40, 0.50}, 0.33166247903554, false},
	}

	for i, test := range tests {
		if rms := rmsF64Samples(test.samples); rms != test.result {
			// If expected result is NaN, continue
			if math.IsNaN(rms) && test.isNaN {
				continue
			}

			t.Fatalf("[%02d] unexpected result: %v != %v", i, rms, test.result)
		}
	}
}
