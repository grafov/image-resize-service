package main

import (
	"github.com/nfnt/resize"

	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"net/http"
	"net/url"
	"strconv"
)

// Useful defaults and limits that could be moved to the config in
// production service.
const (
	minSize         = 32
	maxSize         = 8192
	jpegQuality     = 92
	resizeAlgorithm = resize.Bilinear
)

func handleResizeRequest(w http.ResponseWriter, r *http.Request) {
	var (
		imageURL      string
		width, height uint64
		err           error
	)
	if r.Method != "GET" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	if imageURL, width, height, err = parseParams(r); err != nil {
		http.Error(w, fmt.Sprintf("400 request error: %s", err), http.StatusBadRequest)
		return
	}
	var (
		srcImage image.Image
	)
	if srcImage, err = loadURL(imageURL); err != nil {
		http.Error(w, fmt.Sprintf("426 image loading error: %s", err), http.StatusFailedDependency)
		return
	}
	resizedImage := resize.Resize(uint(width), uint(height), srcImage, resizeAlgorithm)
	opts := jpeg.Options{Quality: jpegQuality}
	// Encoding process below could return error but the main cause of
	// errors here are transport errors due broken connection on
	// client side. They could be ignored safely. Image of very big
	// size also could trigger error but this value far beyond the
	// limits that set by `maxSize` in our code. Maybe for the
	// exceptional safety it is better log error here.
	jpeg.Encode(w, resizedImage, &opts)
}

// When we return more than two values from the function it is good
// choice for clarity let them names in function declaration. It is
// better not return many values of course.
func parseParams(r *http.Request) (imageURL string, width, height uint64, err error) {
	var args url.Values
	if args, err = url.ParseQuery(r.URL.RawQuery); err != nil {
		return
	}
	if args.Get("url") == "" {
		err = errors.New("non empty `url` parameter is mandatory")
		return
	}
	if args.Get("w") == "" {
		err = errors.New("non empty `w` parameter is mandatory")
		return
	}
	if args.Get("h") == "" {
		err = errors.New("non empty `h` parameter is mandatory")
		return
	}
	imageURL = args.Get("url")
	if width, err = strconv.ParseUint(args.Get("w"), 10, 64); err != nil {
		return
	}
	if width > 0 && width < minSize || width > maxSize {
		err = errors.New("width value is out of limit")
		return
	}
	if height, err = strconv.ParseUint(args.Get("h"), 10, 64); err != nil {
		return
	}
	if width > 0 && height < minSize || height > maxSize {
		err = errors.New("height value is out of limit")
		return
	}
	if width == 0 && height == 0 {
		err = errors.New("either width or height should be greater than zero")
		return
	}
	return
}

// Loads data from URL and try to convert it to JPEG.
func loadURL(imageURL string) (image.Image, error) {
	resp, err := http.Get(imageURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return jpeg.Decode(resp.Body)
}
