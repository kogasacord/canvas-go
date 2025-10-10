package styles

import (
	"fmt"
	"image"
	"image/draw"
	"image/gif"
	"image/png"
	"log"
	"os"
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

	resizer := gift.New(gift.Resize(0, src.Config.Height, gift.LinearResampling))
	width := int(float32(src.Config.Height) * 1.77778);
	screenResolution := image.Rect(0, 0, width, src.Config.Height);

    gifQuantizer := utils.NewOctreeQuantizer();
	gifQuantizer.AddColorsFromGIF(src); // COSTLY.
	colorPalette := utils.ConvertToColorPalette(gifQuantizer.MakePalette(256));

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

		regularImage := image.NewPaletted(screenResolution, colorPalette);

		mainImg := image.NewRGBA(screenResolution);
		draw.Draw(mainImg, screenResolution, img, image.Pt(0, 0), draw.Src); // fast.
		draw.Draw(mainImg, screenResolution, resizedOverlayImage, image.Pt(0, 0), draw.Over);
		// alpha-blends correctly ^^

		bounds := mainImg.Bounds();

		// convert this into a function and make a dithering function for the gradient.
		
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				i := y*mainImg.Stride + x*4
				r := mainImg.Pix[i+0]
				g := mainImg.Pix[i+1]
				b := mainImg.Pix[i+2]
				a := mainImg.Pix[i+3]
				if a == 0 {
					continue;
				}
				/*
				if a < 200 {
					fmt.Printf("x: %d, y: %d, (%d, %d, %d), a: %d\n", x, y, r, g, b, a);
				}
				*/

				color := utils.NewColor(int(r), int(g), int(b), int(a));
				index := gifQuantizer.GetPaletteIndex(color);
				if disposal == gif.DisposalPrevious {
					i := reusedImage.PixOffset(x, y);
					reusedImage.Pix[i] = uint8(index);
				} else {
					regularImage.SetColorIndex(x, y, uint8(index));
				}
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

func encodeOverlayImage(img image.Image) {
	f, err := os.Create("overlayimage.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err = png.Encode(f, img); err != nil {
		log.Printf("failed to encode: %v", err)
	}
}
