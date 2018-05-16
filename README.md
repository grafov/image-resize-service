# Example of Image Resizing Service in Go language

The code was written as the sample for the job interview. The code
demonstrates:
* using of HTTP-server and HTTP-client with Go sldlib
* inmem caching of the service results
* simple manipulations with JPEG-image format with Go stdlib
* using of external library
  [github.com/nfnt/resize](https://github.com/nfnt/resize) for image
  resizing in native Go

The service has single HTTP-handler `GET /resize` that accepts `url`
parameter that should pointe to a resource with JPEG-image. It loads
the image, resizes it as requested with `w` and `h` parameters and
returns the result.

The code is short and clean but verbosely commented so you could use
it for studying topics of Go programming related for handling HTTP
requests and simple image processing. If you really looking for fully
functional image resizing service in Go follow the links:

* https://github.com/thoas/picfit
* https://github.com/willnorris/imageproxy

The benchmarks below useful for wise choising of the image
manipulation library:

https://github.com/fawick/speedtest-resize

I've choosen the image lib not wisely :) The `github.com/nfnt/resize`
is not a fastest and it is unmaintained for awhile so for production
uses it is better to pick up something different.

This code of this image resizer is public domain.
