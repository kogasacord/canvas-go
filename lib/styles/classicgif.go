package styles

import (
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"log"
	"time"

	"canvas/lib/utils"

	"github.com/disintegration/gift"
	// "github.com/fogleman/gg"
	"golang.org/x/image/font"
)

func ModifyClassicGif(
	src *gif.GIF, 
	font *font.Face, 
	small_font *font.Face, 
	text string, 
	author string, 
	gradient *image.Image,
) *gif.GIF {
	start := time.Now();

	newGif := &gif.GIF{};

	resizer := gift.New(gift.Resize(0, src.Config.Height, gift.LinearResampling));
	width := int(float32(src.Config.Height) * 1.77778);
	screenResolution := image.Rect(0, 0, width, src.Config.Height);

    gifQuantizer := utils.NewOctreeQuantizer();
	gifQuantizer.AddColorsFromGIF(src); // COSTLY.
	colorPalette := utils.ConvertToColorPalette(gifQuantizer.MakePalette(255));

	transparentColor := color.RGBA{0, 0, 0, 0};
	colorPalette = append(colorPalette, transparentColor);
	transparentIndex := len(colorPalette) - 1;
	// put a transparent color palette in the palette
	// transparentIndex gets pushed in the image when the alpha is below 255.

	gradientAssumedSize := image.Rect(0, 0, 1280, 720);
	tempEmptyImage := image.NewRGBA(gradientAssumedSize);
	overlayCanvas := composeClassicImage(tempEmptyImage, *gradient, *font, *small_font, gradientAssumedSize, text, author, 400, 9);

	resizedOverlayImage := image.NewRGBA(screenResolution);
	resizer.Draw(resizedOverlayImage, overlayCanvas.Image());
	// ^^ maintains transparency!

	reusedImage := image.NewPaletted(screenResolution, colorPalette);
	for i := 0; i < len(src.Image); i++ {
		img := src.Image[i];
		delay := src.Delay[i];
		disposal := src.Disposal[i];
		ditherBuffer := utils.NewDitherBuffer(screenResolution.Max.X, screenResolution.Max.Y);

		regularImage := image.NewPaletted(screenResolution, colorPalette);

		mainImg := image.NewRGBA(screenResolution);
		draw.Draw(mainImg, screenResolution, img, image.Pt(0, 0), draw.Src); // fast.
		draw.Draw(mainImg, screenResolution, resizedOverlayImage, image.Pt(0, 0), draw.Over);
		// alpha-blends correctly ^^

		// ditherErrorBuffer := image.NewRGBA(screenResolution);
		
		for y := screenResolution.Min.Y; y < screenResolution.Max.Y; y++ {
			for x := screenResolution.Min.X; x < screenResolution.Max.X; x++ {
				rgbaIndex := y*mainImg.Stride + x*4
				errorIndex := (y-screenResolution.Min.Y)*screenResolution.Dx() + (x-screenResolution.Min.X)
				r := mainImg.Pix[rgbaIndex+0]
				g := mainImg.Pix[rgbaIndex+1]
				b := mainImg.Pix[rgbaIndex+2]
				a := mainImg.Pix[rgbaIndex+3]
				color := utils.NewColor(int(r), int(g), int(b), int(a));

				bufferError := ditherBuffer.GetErrorIndex(max(errorIndex - 1, 0));
				appliedErrorColor := ditherBuffer.ApplyErrorToColor(bufferError, color);

				var paletteIndex int
				if a < 254 {
					paletteIndex = transparentIndex;
				} else {
					paletteIndex = gifQuantizer.GetPaletteIndex(color);
				}
				palettedColor := utils.ConvertToColor(colorPalette[paletteIndex]);

				ditherBuffer.DiffusePixelWithFloydSteinberg(x, y, appliedErrorColor, palettedColor);

				reusedImage.SetColorIndex(x, y, uint8(paletteIndex));
				regularImage.SetColorIndex(x, y, uint8(paletteIndex));
			}
		}

		if disposal == gif.DisposalPrevious {
			newGif.Image = append(newGif.Image, reusedImage);
		} else {
			newGif.Image = append(newGif.Image, regularImage);
		}
		newGif.Delay = append(newGif.Delay, delay);
		newGif.Disposal = append(newGif.Disposal, disposal);
	}

	newGif.LoopCount = src.LoopCount;
	newGif.Config.Height = src.Config.Height;
	newGif.Config.Width = width;
	newGif.Config.ColorModel = src.Config.ColorModel;
	newGif.BackgroundIndex = src.BackgroundIndex;

	elapsed := time.Since(start);
	log.Printf("classicgif v1 took %s\n", elapsed);

	return newGif;
}

