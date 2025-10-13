
package utils

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

func (buf *DitherBuffer) ApplyErrorToColor(errColor Color, color Color) Color {
	return Color {
		Red:   min(max(errColor.Red + color.Red, 0), 255),
		Green: min(max(errColor.Green + color.Green, 0), 255),
		Blue:  min(max(errColor.Blue + color.Blue, 0), 255),
		Alpha: min(max(errColor.Alpha + color.Alpha, 0), 255),
	}
}

func (buf *DitherBuffer) DiffusePixelWithFloydSteinberg(x, y int, source Color, corrected Color)  {
	errorBufferLength := len(buf.errorBuffer) - 1;
	errorIndex := min((buf.w * y) + x, errorBufferLength);
	nextRowIndex := min((buf.w * (y + 1)) + x, errorBufferLength);

	errColor := source.Subtract(corrected);

	frontCurrRowItem := &buf.errorBuffer[min(errorIndex + 1, errorBufferLength)];
	behindNextRowItem := &buf.errorBuffer[min(nextRowIndex - 1, errorBufferLength)];
	nextRowItem := &buf.errorBuffer[nextRowIndex];
	frontNextRowItem := &buf.errorBuffer[min(nextRowIndex + 1, errorBufferLength)];

	frontCurrRowItem.AddMutate(errColor.MultiplyByNum(7).DivideByNum(16));
	frontNextRowItem.AddMutate(errColor.DivideByNum(16));
	nextRowItem.AddMutate(errColor.MultiplyByNum(5).DivideByNum(16));
	behindNextRowItem.AddMutate(errColor.MultiplyByNum(3).DivideByNum(16));
}



