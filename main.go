package main

import (
	// "image"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"runtime/pprof"

	"canvas/lib/styles"
	"canvas/lib/utils"

	"github.com/fogleman/gg"

	"image/gif"
	_ "image/png"
)

var gradient = utils.OpenImage("./images/quote/qgradient.png");
var big_classic_font, _ = gg.LoadFontFace("./fonts/Mirador-SemiBold.ttf", 25 * 2);
var small_classic_font, _ = gg.LoadFontFace("./fonts/Mirador-BookItalic.ttf", 15 * 2);

var big_classicgif_font, _ = gg.LoadFontFace("./fonts/Mirador-SemiBold.ttf", 25);
var small_classicgif_font, _ = gg.LoadFontFace("./fonts/Mirador-BookItalic.ttf", 15);

var gifminimalist_font, _ = gg.LoadFontFace("./fonts/Lora-Italic.ttf", 25)



type QuoteMeta struct {
	AvatarUrl string `json:"avatar_url"`
	Author string `json:"author"`
	Text string `json:"text"`
}
type ImageQuoteMeta struct {
	AvatarUrl string `json:"avatar_url"`
	ImageUrl string `json:"image_url"`
	Author string `json:"author"`
	Text string `json:"text"`
}

func sendFramedImageQuote(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := io.ReadAll(r.Body);
	var meta QuoteMeta;

	if err := json.Unmarshal(reqBody, &meta); err != nil {
		http.Error(w, "Failed to parse metadata.", http.StatusBadRequest);
		return;
	}
	/*
	img, err := GetImageFromUrl(meta.Url);
	if err != nil {
		http.Error(w, "Can't get image from URL. " + err.Error(), http.StatusBadRequest);
		return;
	}
	*/
}

func sendMinimalistQuote(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := io.ReadAll(r.Body);
	var meta QuoteMeta;

	if err := json.Unmarshal(reqBody, &meta); err != nil {
		http.Error(w, "Failed to parse metadata.", http.StatusBadRequest);
		return;
	}
	img, err := utils.GetImageFromURL(meta.AvatarUrl);
	if err != nil {
		http.Error(w, "Can't get image from URL. " + err.Error(), http.StatusBadRequest);
		return;
	}
	imgData, err := styles.ModifyMinimalistImage(img, &small_classic_font, meta.Text);
	if err != nil {
		http.Error(w, "minimalist image error: " + err.Error(), http.StatusBadRequest);
		return;
	}
	w.Header().Set("Content-Type", "image/png");
	imgData.EncodePNG(w);
}


// todo: combine classic quote and gif quote together.
func sendClassicQuote(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := io.ReadAll(r.Body);
	var meta QuoteMeta;

	if err := json.Unmarshal(reqBody, &meta); err != nil {
		http.Error(w, "Failed to parse metadata.", http.StatusBadRequest);
		return;
	}
	img, err := utils.GetImageFromURL(meta.AvatarUrl);
	if err != nil {
		http.Error(w, "Can't get image from URL. " + err.Error(), http.StatusBadRequest);
		return;
	}
	imgData := styles.ModifyClassicImage(meta.Text, meta.Author, img, gradient, &big_classic_font, &small_classic_font);
	w.Header().Set("Content-Type", "image/png");
	imgData.EncodePNG(w);
}

func sendClassicGifQuote(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := io.ReadAll(r.Body);
	var meta QuoteMeta;

	if err := json.Unmarshal(reqBody, &meta); err != nil {
		http.Error(w, "Failed to parse metadata.", http.StatusBadRequest);
		return;
	}

	// get gif instead.
	decoded_gif, err := utils.GetGifFromURL(meta.AvatarUrl);
	if err != nil {
		http.Error(w, "Can't get image from URL. " + err.Error(), http.StatusBadRequest);
		return;
	}

	gifData := styles.ModifyClassicGif(decoded_gif, &big_classic_font, &small_classic_font, meta.Text, meta.Author, &gradient);
	w.Header().Set("Content-Type", "image/gif");
	gif.EncodeAll(w, gifData);
}


func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Pong!"));
}

func testGIFCompose()  {
	testGIF := utils.OpenGIF("./images/test.gif");
	styles.ModifyClassicGif(
		testGIF, 
		&big_classic_font, 
		&small_classic_font, 
		"bless up", 
		"- tempt", 
		&gradient,
	);
}

func profileCPU() {
	testGIFCompose();
}

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file");
func main() {
	flag.Parse();
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile);
		if err != nil {
			log.Fatal(err);
		}
		pprof.StartCPUProfile(f);
		defer pprof.StopCPUProfile();

		println("Profiling!");
		profileCPU();

		return;
	}

	http.HandleFunc("/ping", ping);
	http.HandleFunc("/quote", sendClassicQuote);
	http.HandleFunc("/quotegif", sendClassicGifQuote);
	http.HandleFunc("/quotemini", sendMinimalistQuote);
	http.HandleFunc("/quoteframe", sendFramedImageQuote);
	println("Started server on localhost:8080");
	http.ListenAndServe(":8080", nil);
}
