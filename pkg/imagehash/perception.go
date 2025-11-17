package imagehash

import (
	"errors"
	"image"
	"math"
	"sync"

	"github.com/corona10/goimagehash"
	"github.com/corona10/goimagehash/etcs"
	"github.com/corona10/goimagehash/transforms"
)

const (
	sampleSize = 64
	hashWidth  = 8
	hashHeight = 8
)

var pixelPool = sync.Pool{
	New: func() interface{} {
		p := make([]float64, sampleSize*sampleSize)
		return &p
	},
}

// PerceptionHash calculates a perceptual hash for the provided image while avoiding
// the heavy intermediate allocations required by github.com/nfnt/resize as implemented
// in goimagehash. Essentially, this is just a small reimplemenation.
func PerceptionHash(img image.Image) (string, error) {
	if img == nil {
		return "", errors.New("image is nil")
	}

	bounds := img.Bounds()
	if bounds.Dx() == 0 || bounds.Dy() == 0 {
		return "", errors.New("image has invalid bounds")
	}

	pixelsPtr := pixelPool.Get().(*[]float64)
	defer pixelPool.Put(pixelsPtr)

	fillDownsampledGrayscale(img, *pixelsPtr)

	transforms.DCT2DFast64(pixelsPtr)
	flattens := transforms.FlattenPixelsFast64(*pixelsPtr, hashWidth, hashHeight)
	median := etcs.MedianOfPixelsFast64(flattens)

	var hash uint64
	for idx, p := range flattens {
		if p > median {
			shift := uint(len(flattens) - idx - 1)
			hash |= 1 << shift
		}
	}

	return goimagehash.NewImageHash(hash, goimagehash.PHash).ToString(), nil
}

func fillDownsampledGrayscale(src image.Image, dst []float64) {
	bounds := src.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	scaleX := float64(srcWidth) / float64(sampleSize)
	scaleY := float64(srcHeight) / float64(sampleSize)

	for y := range sampleSize {
		sy := bounds.Min.Y + sampleCoordinate(scaleY, srcHeight, y)
		for x := range sampleSize {
			sx := bounds.Min.X + sampleCoordinate(scaleX, srcWidth, x)
			r, g, b, _ := src.At(sx, sy).RGBA()
			dst[(y*sampleSize)+x] = toGray(r, g, b)
		}
	}
}

func sampleCoordinate(scale float64, max int, pos int) int {
	if max <= 1 {
		return 0
	}

	value := int(math.Floor((float64(pos) + 0.5) * scale))
	if value < 0 {
		return 0
	}
	if value >= max {
		return max - 1
	}

	return value
}

func toGray(r, g, b uint32) float64 {
	return 0.299*float64(r/257) + 0.587*float64(g/257) + 0.114*float64(b/256)
}
