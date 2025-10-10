package styles

import (
	"image"
	"image/gif"

	"golang.org/x/image/font"

	"github.com/fogleman/gg"

	"canvas/lib/utils"
)

type GifFrame struct {
	palettedImage *image.Paletted
	delay int
	index int
}

func ModifyMinimalistGif(src *gif.GIF, font *font.Face, text string) *gif.GIF {
	newGif := &gif.GIF{};

    quantizer := utils.NewOctreeQuantizer()
	quantizer.AddColorsFromGIF(src);
    colorCount := 256; // colors. 256 before.
	palette := quantizer.MakePalette(colorCount)
	colorPalette := utils.ConvertToColorPalette(palette);

	average_luminosity, _ := utils.GetAverageBrightnessOfPalettedImage(src.Image[0], src.Config.Width, src.Config.Height);
	screenResolution := image.Rect(0, 0, src.Config.Width, src.Config.Height);

	// the reused image.
	reusedImage := image.NewPaletted(screenResolution, colorPalette);

	for i := 0; i < len(src.Image); i++ {
		img := src.Image[i];
		delay := src.Delay[i];
		disposal := src.Disposal[i];

		regularImage := image.NewPaletted(screenResolution, colorPalette);

		dc := composeMinimalistFrameGif(img, *font, text, screenResolution, average_luminosity);
		dcImg := dc.Image().(*image.RGBA);
		bounds := dcImg.Bounds();

		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				i := y*dcImg.Stride + x*4
				// Extract the RGBA values directly from the Pix slice
				r := dcImg.Pix[i+0]
				g := dcImg.Pix[i+1]
				b := dcImg.Pix[i+2]
				a := dcImg.Pix[i+3]
				if a == 0 {
					continue;
				}

				color := utils.NewColor(int(r), int(g), int(b), int(a));
				index := quantizer.GetPaletteIndex(color);
				// if disposalNone, then we modify the reusedImage again.
				// when more frames gets parsed, this reusedImage gets populated more.
				if disposal == gif.DisposalNone {
					i := reusedImage.PixOffset(x, y);
					reusedImage.Pix[i] = uint8(index);
				} else {
					// regularImage is made in this loop, meaning it'll reset every frame.
					regularImage.SetColorIndex(x, y, uint8(index));
				}
			}
		}

		if disposal == gif.DisposalNone {
			copiedReusedImage := image.NewPaletted(screenResolution, colorPalette)
			copy(copiedReusedImage.Pix, reusedImage.Pix)
			newGif.Image = append(newGif.Image, copiedReusedImage);
		} else {
			newGif.Image = append(newGif.Image, regularImage);
		}
		newGif.Delay = append(newGif.Delay, delay);
		newGif.Disposal = append(newGif.Disposal, 1);
	}

	newGif.LoopCount = src.LoopCount;
	newGif.Config.Height = src.Config.Height;
	newGif.Config.Width = src.Config.Width;
	newGif.Config.ColorModel = src.Config.ColorModel;
	newGif.BackgroundIndex = src.BackgroundIndex;

	return newGif;
}

// this automatically adapts to the image resolutions
func composeMinimalistFrameGif(
	img *image.Paletted,
	font font.Face,
	text string, 
	resolution image.Rectangle,
	average_luminosity uint32,
) *gg.Context {
	gifWidth := resolution.Max.X;
	gifHeight := resolution.Max.Y;

	imgXOrigin := img.Bounds().Min.X;
	imgYOrigin := img.Bounds().Min.Y;
	imgWidth := img.Bounds().Max.X;
	imgHeight := img.Bounds().Max.Y;

	var dc *gg.Context;
	if 0 != imgXOrigin || 0 != imgYOrigin || imgWidth != gifWidth || imgHeight != gifHeight {
		dc = gg.NewContext(gifWidth, gifHeight);
		dc.DrawImage(img, 0, 0); // significantly slower.
	} else {
		dc = gg.NewContextForImage(img);
	}

	r, g, b := 255, 255, 255;
	if average_luminosity > 150 {
		r, g, b = 0, 0, 0;
	}
	dc.SetFontFace(font);

	offset := 10.0;
	dc.SetRGBA255(r, g, b, 255);
	dc.DrawRectangle(0 + offset, 0 + offset, float64(gifWidth) - (10 + offset), float64(gifHeight) - (10 + offset));
	dc.SetLineWidth(1);
	dc.Stroke();

	dc.SetRGBA255(r, g, b, 255);
	dc.DrawStringWrapped(
		text,
		float64(dc.Width()), // x
		float64(dc.Height()) / 2, // y
		1.1, // ax (anchor x)
		0.5, // ay (anchor y)
		float64(dc.Width()) / 2, // width
		1.2, // line spacing
		gg.AlignRight,
	);
	return dc;
}
