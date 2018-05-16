package main

import (
	"github.com/coocood/freecache"
	"github.com/stretchr/testify/assert"

	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

const (
	uniqImageSize = minSize + 2 // for avoiding caching during the tests, all other requests use minSize+1
)

var (
	goodImageURL   string
	brokenImageURL string
)

// Copypasted main() for CLI parsing, init cache and
// webserver. This is setup for blackbox-like testing for the
// handlers.
func TestMain(m *testing.M) {
	flag.StringVar(&hostPort, "listen-at", "localhost:8080", "listen for HTTP requests at host:port")
	flag.Parse()

	goodImageURL = "http://" + hostPort + "/static/l_hires.jpg"
	brokenImageURL = "http://" + hostPort + "/static/nonjpeg.jpg"

	cache = freecache.NewCache(cacheSize)

	// Testing only handler with pictures samples:
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("testdata"))))

	// Handlers that should be tested:
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleRootRequest(w, r)
	})
	http.HandleFunc("/resize", func(w http.ResponseWriter, r *http.Request) {
		handleResizeRequest(w, r)
	})
	go func() { http.ListenAndServe(hostPort, nil) }()
	time.Sleep(100 * time.Millisecond)
	ret := m.Run()
	os.Exit(ret)
}

func TestSampleJpegImageExists(t *testing.T) {
	resp, err := http.Get(goodImageURL)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestSampleNonJpegImageExists(t *testing.T) {
	resp, err := http.Get(brokenImageURL)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetNotExistedHandler(t *testing.T) {
	resp, err := http.Get("http://" + hostPort + "/should-be-not-found")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetRoot(t *testing.T) {
	resp, err := http.Get("http://" + hostPort)
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}

	assert.Contains(t, string(data), fmt.Sprintf("Sample Image Resizer ver. %s", version))
}

func TestPutResize_WrongMethod(t *testing.T) {
	client := &http.Client{}
	req, err := http.NewRequest("PUT", "http://"+hostPort+"/resize", nil)
	if err != nil {
		t.Error(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

func TestGetResize_NoArgs(t *testing.T) {
	resp, err := http.Get("http://" + hostPort + "/resize")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetResize_TotallyWrongArgs(t *testing.T) {
	resp, err := http.Get("http://" + hostPort + "/resize?username=xxx&password=yyy")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetResize_NoSizes(t *testing.T) {
	resp, err := http.Get("http://" + hostPort + "/resize?url=xxx.jpeg")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetResize_NoHeight(t *testing.T) {
	resp, err := http.Get(fmt.Sprintf("http://%s/resize?url=%s&width=%d", hostPort, goodImageURL, minSize+1))
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetResize_NoWidth(t *testing.T) {
	resp, err := http.Get(fmt.Sprintf("http://%s/resize?url=%s&height=%d", hostPort, goodImageURL, minSize+1))
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetResize_SourceImageDoesntExists(t *testing.T) {
	resp, err := http.Get(fmt.Sprintf("http://%s/resize?url=notexisted.jpeg&height=%d&width=%d", hostPort, minSize+1, minSize+1))
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, http.StatusFailedDependency, resp.StatusCode)
}

func TestGetResize_Ok(t *testing.T) {
	resp, err := http.Get(fmt.Sprintf("http://%s/resize?url=%s&width=%d&height=%d", hostPort, goodImageURL, minSize+1, minSize+1))
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetResize_NonJpeg(t *testing.T) {
	resp, err := http.Get(fmt.Sprintf("http://%s/resize?url=%s&width=%d&height=%d", hostPort, brokenImageURL, minSize+1, minSize+1))
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, http.StatusFailedDependency, resp.StatusCode)
}

func TestGetResize_ExceedWidthLimit(t *testing.T) {
	resp, err := http.Get(fmt.Sprintf("http://%s/resize?url=%s&height=%d&width=%d", hostPort, goodImageURL, minSize+1, maxSize+1))
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetResize_ExceedHeightLimit(t *testing.T) {
	resp, err := http.Get(fmt.Sprintf("http://%s/resize?url=%s&height=%d&width=%d", hostPort, goodImageURL, maxSize+1, minSize+1))
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetResize_KeepRatioByWidth(t *testing.T) {
	resp, err := http.Get(fmt.Sprintf("http://%s/resize?url=%s&width=0&height=%d", hostPort, goodImageURL, minSize+1))
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetResize_KeepRatioByHeight(t *testing.T) {
	resp, err := http.Get(fmt.Sprintf("http://%s/resize?url=%s&width=%d&height=0", hostPort, goodImageURL, minSize+1))
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetResize_CheckForEtag(t *testing.T) {
	fullreq := fmt.Sprintf("http://%s/resize?url=%s&width=%d&height=0", hostPort, goodImageURL, uniqImageSize)
	resp, err := http.Get(fullreq)
	if err != nil {
		t.Error(err)
	}
	w := httptest.NewRecorder()
	ok := useClientCache(w, resp.Request)
	writtenResp := w.Result()

	assert.False(t, ok)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, resp.Header.Get("Etag"), writtenResp.Header.Get("Etag"))
}
