package utils

import (
	"image"
	"image/color"
	"image/png"
	"os"
)

type DitherBuffer struct {
	errorBuffer []Color
	w int
	h int
}

func NewDitherBuffer(w, h int) DitherBuffer {
	errorBuffer := make([]Color, (w * h));
	return DitherBuffer { errorBuffer, w, h };
}

func (buf *DitherBuffer) GetErrorIndex(i int) Color {
	if i >= buf.w * buf.h {
		println("went over! ", i, buf.w * buf.h)
	}
	return buf.errorBuffer[i];
}

func absInt(x int) int {
    if x < 0 { return -x }
    return x
}

func clampInt(v, lo, hi int) int {
    if v < lo { return lo }
    if v > hi { return hi }
    return v
}

func (buf *DitherBuffer) DumpErrorBuffer(filename string) {
    w := buf.w
    h := buf.h
    img := image.NewGray(image.Rect(0, 0, w, h))

    // find max abs component to normalize later
    maxVal := 0
    for _, c := range buf.errorBuffer {
        for _, v := range []int{c.Red, c.Green, c.Blue} {
            if absInt(v) > maxVal {
                maxVal = absInt(v)
            }
        }
    }
    if maxVal == 0 {
        maxVal = 1
    }

    for y := 0; y < h; y++ {
        for x := 0; x < w; x++ {
            c := buf.errorBuffer[y*w+x]
            // average RGB, normalized into 0-255 with midpoint=128
            val := (c.Red + c.Green + c.Blue) / 3
            scaled := 128 + (val * 127 / maxVal) // 128 = zero point
			scaled = clampInt(scaled, 0, 255)
            img.SetGray(x, y, color.Gray{uint8(scaled)})
        }
    }

    outFile, _ := os.Create(filename)
    defer outFile.Close()
    png.Encode(outFile, img)
}

func (buf *DitherBuffer) ApplyErrorToColor(errColor Color, color Color) Color {
	return Color {
		Red:   min(max(errColor.Red + color.Red, 0), 255),
		Green: min(max(errColor.Green + color.Green, 0), 255),
		Blue:  min(max(errColor.Blue + color.Blue, 0), 255),
		Alpha: min(max(errColor.Alpha + color.Alpha, 0), 255),
	}
}

// replace with ordered dithering
func (buf *DitherBuffer) DiffusePixelWithFloydSteinberg(x, y int, source Color, corrected Color)  {
	errorBufferLength := len(buf.errorBuffer) - 1;
	errorIndex := min((buf.w * y) + x, errorBufferLength);
	nextRowIndex := min((buf.w * (y + 1)) + x, errorBufferLength);

	errColor := source.Subtract(corrected);

	frontCurrRowItem := &(buf.errorBuffer[min(errorIndex + 1, errorBufferLength)]);
	behindNextRowItem := &(buf.errorBuffer[min(nextRowIndex - 1, errorBufferLength)]);
	nextRowItem := &(buf.errorBuffer[nextRowIndex]);
	frontNextRowItem := &(buf.errorBuffer[min(nextRowIndex + 1, errorBufferLength)]);

	frontCurrRowItem.AddMutate(errColor.MultiplyByNum(7).DivideByNum(16));
	frontNextRowItem.AddMutate(errColor.DivideByNum(16));
	nextRowItem.AddMutate(errColor.MultiplyByNum(5).DivideByNum(16));
	behindNextRowItem.AddMutate(errColor.MultiplyByNum(3).DivideByNum(16));
}



