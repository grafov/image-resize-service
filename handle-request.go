package main

import (
	"github.com/nfnt/resize"

	"bytes"
	"errors"
	"fmt"
	"hash/fnv"
	"image"
	"image/jpeg"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Useful defaults and limits that could be moved to the config in
// production service.
const (
	minSize         = 32
	maxSize         = 8192
	jpegQuality     = 92
	resizeAlgorithm = resize.Bilinear
	cachingDuration = 1 * time.Hour
)

// Implements handler for `/resize`. Moved out of main() for code clarity.
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
	if useClientCache(w, r) {
		return
	}
	if useServerCache(w, imageURL, width, height) {
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
	// client side. They could be ignored safely we just skip caching
	// of partial results.
	buf := new(bytes.Buffer)
	if err = jpeg.Encode(buf, resizedImage, &opts); err != nil {
		return
	}
	w.Write(buf.Bytes())
	cache.Set(formatCacheKey(imageURL, width, height), buf.Bytes(), int(cachingDuration.Seconds()))
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
	if args.Get("width") == "" {
		err = errors.New("non empty `width` parameter is mandatory")
		return
	}
	if args.Get("height") == "" {
		err = errors.New("non empty `height` parameter is mandatory")
		return
	}
	imageURL = args.Get("url")
	if width, err = strconv.ParseUint(args.Get("width"), 10, 64); err != nil {
		return
	}
	if width > 0 && width < minSize || width > maxSize {
		err = errors.New("width value is out of limit")
		return
	}
	if height, err = strconv.ParseUint(args.Get("height"), 10, 64); err != nil {
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

// Use client cache where possible.
func useClientCache(w http.ResponseWriter, r *http.Request) bool {
	const sep = "X"
	hash := fnv.New64()
	hash.Write([]byte(r.URL.RawQuery))
	// Etag value has two parts: 1) hash based on URL string 2) timestamp.
	etagHash := strconv.FormatUint(hash.Sum64(), 10)
	etagTs := strconv.FormatInt(time.Now().Unix(), 10)

	if match := r.Header.Get("If-None-Match"); match != "" {
		partsOfEtag := strings.SplitN(match, sep, 2)
		// Check for etag has two parts as declared above and URL-hash is equal.
		if len(partsOfEtag) == 2 && partsOfEtag[0] == etagHash {
			// Second part of etag should be timestamp so we can check
			// duration since previuous call for this URL.
			clientTs, _ := strconv.ParseInt(partsOfEtag[1], 10, 64)
			if time.Since(time.Unix(clientTs, 0)) < cachingDuration {
				w.WriteHeader(http.StatusNotModified)
				return true
			}
		}
	}

	w.Header().Set("Etag", etagHash+sep+etagTs)
	return false
}

// Inmem cache with LRU. Keys expired after `cachingDuration`.
func useServerCache(w http.ResponseWriter, url string, width, height uint64) bool {
	var (
		data []byte
		err  error
	)
	if data, err = cache.Get(formatCacheKey(url, width, height)); err != nil {
		return false
	}
	w.Write(data)
	return true
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

// Helper for making the key for caching in single place.
func formatCacheKey(url string, width, height uint64) []byte {
	buf := new(bytes.Buffer)
	buf.WriteString(url)
	buf.WriteRune(':')
	buf.WriteString(strconv.FormatUint(width, 10))
	buf.WriteRune(':')
	buf.WriteString(strconv.FormatUint(height, 10))
	return buf.Bytes()
}
