# Example of Image Resizing Service in Go language

The code was written as the sample for the job interview. The code
demonstrates:
* using of HTTP-server and HTTP-client with Go sldlib
* advice client for cache results with using Etag
* inmem caching of the results with limited cache size and LRU [github.com/coocood/freecache](https://github.com/coocood/freecache)
* JPEG resizing with library
  [github.com/nfnt/resize](https://github.com/nfnt/resize)

The service has single HTTP-handler `GET /resize` that accepts `url`
parameter that should pointe to a resource with JPEG-image. It loads
the image, resizes it as requested with `w` and `h` parameters and
returns the result.

The code is short and clean but verbosely commented so you could use
it for studying topics of Go programming related for handling HTTP
requests and simple image processing.

The benchmarks below useful for wise choising of the image
manipulation library:

https://github.com/fawick/speedtest-resize

I've choosen the image lib not wisely :) The `github.com/nfnt/resize`
is not a fastest and it is unmaintained for awhile so for production
uses it is better to pick up something different.

This code of this image resizer is public domain. I have no plans to
add more features or maintain it. If you really looking for fully
functional image resizing service in Go follow the links:

* https://github.com/thoas/picfit
* https://github.com/willnorris/imageproxy

## Some hints for Go newcomers who could follow this sample code

* With small codebase don't split the code by packages early. It is
  maybe good practice in Python and mandatory in Java but Go doesn't
  require you do what you don't need. The small code looks good in
  `main` package.
* Keep the `main()` as small as possible. Move all to functions in
  other files.
* Use `TestMain()` (was introduced in Go 1.4 but I rarely see it in
  the tests) when you want make blackbox-like testing. `TestMain()`
  will help you set up testing environment like if you `go build` and
  run you code.
* Even with small codebase put your dependencies to `vendor`
  package. First way is exclude this folder from version control but
  don't forget to clean it regulary for avoid possible local changes
  inside `vendor`. Second way is include `vendor` into version control
  so VCS will help your check that changed after each dependency
  update. I prefer the second way for the code with small and medium
  number of dependencies (`govendor` utility does it this way). For
  large number of dependencies is better keep them out of version
  control.
