package utils

import (
	"bufio"
	"fmt"
	"image"
	"image/gif"
	"net/http"
	"os"
)

func OpenImage(fp string) image.Image {
	image_file, err := os.Open(fp);
	if err != nil {
		panic(err);
	}
	im := bufio.NewReader(image_file);
	decoded_im, _, err := image.Decode(im);
	if err != nil {
		panic(err);
	}
	if err := image_file.Close(); err != nil {
		panic(err);
	}
	return decoded_im;
}

func OpenGIF(fp string) *gif.GIF {
	gif_file, err := os.Open(fp);
	if err != nil {
		panic(err);
	}
	g := bufio.NewReader(gif_file);
	decoded_gif, err := gif.DecodeAll(g);
	if err != nil {
		panic(err);
	}
	if err := gif_file.Close(); err != nil {
		panic(err);
	}
	return decoded_gif;
}

func GetImageFromURL(url string) (image.Image, error) {
	resp, err := http.Get(url);
	if err != nil {
		return nil, err;
	}
	defer resp.Body.Close();

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status);
	}
	img, _, err := image.Decode(resp.Body);
	if err != nil {
		return nil, err;
	}

	return img, nil;
}

func GetGifFromURL(url string) (*gif.GIF, error) {
	resp, err := http.Get(url);
	if err != nil {
		return nil, err;
	}
	defer resp.Body.Close();

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status);
	}
	decoded_gif, err := gif.DecodeAll(resp.Body);
	if err != nil {
		return nil, err;
	}

	return decoded_gif, nil;

}
