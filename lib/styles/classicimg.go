package styles

import (
	"image"
	"image/color"
	"strings"

	"github.com/disintegration/gift"
	"github.com/fogleman/gg"
	"golang.org/x/image/font"
)

// for use in images.
func ModifyClassicImage(text string, author string, src image.Image, gradient image.Image, font *font.Face, small_font *font.Face) *gg.Context {
	screenResolution := image.Rect(0, 0, 1280, 720);
	resized := image.NewRGBA(screenResolution);
	g := gift.New(
		gift.Resize(720, 0, gift.LanczosResampling), // fastest upscaling.
		gift.Brightness(-30),
	);
	g.Draw(resized, src);

	return composeClassicImage(resized, gradient, *font, *small_font, screenResolution, text, author, 600, 9);
}

// img should be below 720x720 to reap LanczosResampling's speed in upscaling.
// gradient should be exactly 1280x720.
func composeClassicImage(
	img image.Image,
	gradient image.Image,
	font font.Face,
	small_font font.Face,
	resolution image.Rectangle,
	text string,
	author string,
	wrap_width float64,
	line_limit int,
) *gg.Context { // 720p
	text_x := float64(resolution.Max.X) - 40;
	text_y := float64(resolution.Max.Y / 2) - 20; // accounting the author text (20)
	
	dc := gg.NewContext(resolution.Max.X, resolution.Max.Y);
	dc.DrawImage(img, 0, 0);
	dc.DrawImage(gradient, 0, 0);

	dc.SetFontFace(font);
	s := dc.WordWrap(text, wrap_width);
	if len(s) > line_limit {
		s = s[:line_limit];
	}
	stext := strings.Join(s, "\n");

	var total_text_height float64 = 0;
	for _, line := range s {
		_, line_height := dc.MeasureString(line)
		total_text_height += line_height - 23;
	}
	dc.SetColor(color.White);
	dc.DrawStringWrapped(stext, text_x, text_y, 1, 0.5, wrap_width, 1, gg.AlignCenter);

	dc.SetColor(color.White);
	dc.SetFontFace(small_font);
	dc.DrawStringWrapped(author, text_x, text_y + total_text_height, 1, 0, wrap_width / 2, 1, gg.AlignCenter);

	return dc;
}

