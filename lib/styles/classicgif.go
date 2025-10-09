package styles

import (
	"image"
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

	resizer := gift.New(gift.Resize(0, src.Config.Height, gift.LinearResampling))
	width := int(float32(src.Config.Height) * 1.77778);

	// why not use the source gif's palette to save us the work?
	// pre-quantize the overlay image to be in the source gif palette..?
    quantizer := utils.NewOctreeQuantizer()
	utils.AddColorsToQuantizer(quantizer, src);
    colorCount := 256; // colors. 256 before.
	palette := quantizer.MakePalette(colorCount)
	colorPalette := utils.ConvertToColorPalette(palette);
	screenResolution := image.Rect(0, 0, width, src.Config.Height);

	reusedImage := image.NewPaletted(screenResolution, colorPalette);

	// an attempt to render the image 
	gradient_assumed_size := image.Rect(0, 0, 1280, 720);
	temp_empty_image := image.NewPaletted(gradient_assumed_size, colorPalette);
	overlay_canvas := composeClassicImage(temp_empty_image, *gradient, *font, *small_font, gradient_assumed_size, text, author, 400, 9);
	overlay_canvas.SavePNG("./testtest.png");

	overlay_image := image.NewRGBA(screenResolution);
	resizer.Draw(overlay_image, overlay_canvas.Image());
	overlay_quant_image := quantizer.ConvertRGBAToPalettedImage(overlay_canvas.Image().(*image.RGBA), colorPalette);

	for i := 0; i < len(src.Image); i++ {
		img := src.Image[i];
		delay := src.Delay[i];
		disposal := src.Disposal[i];

		regularImage := image.NewPaletted(screenResolution, colorPalette);

		main_img := image.NewRGBA(screenResolution);
		// fast for blitting!
		draw.Draw(main_img, screenResolution, img, image.Pt(0, 0), draw.Src);
		draw.Draw(main_img, screenResolution, overlay_quant_image, image.Pt(0, 0), draw.Over);

		bounds := main_img.Bounds();

		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				i := y*main_img.Stride + x*4
				r := main_img.Pix[i+0]
				g := main_img.Pix[i+1]
				b := main_img.Pix[i+2]
				a := main_img.Pix[i+3]
				if a == 0 {
					continue;
				}

				color := utils.NewColor(int(r), int(g), int(b), int(a));
				index := quantizer.GetPaletteIndex(color);
				if disposal == gif.DisposalNone {
					i := reusedImage.PixOffset(x, y);
					reusedImage.Pix[i] = uint8(index);
				} else {
					regularImage.SetColorIndex(x, y, uint8(index));
				}
			}
		}

		if disposal == gif.DisposalNone {
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
