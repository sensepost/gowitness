<h1 align="center">
  <br>
  <a href="https://github.com/sensepost/objection">
    <img src="images/gowitness-logo.png" alt="objection"></a>
  <br>
  <br>
</h1>

<h4 align="center">A golang, web screenshot utility using Chrome Headless.</h4>
<p align="center">
  <a href="https://twitter.com/leonjza"><img src="https://img.shields.io/badge/Twitter-%40leonjza-blue.svg" alt="@leonjza" height="18"></a>
  <a href="https://goreportcard.com/report/github.com/sensepost/gowitness"><img src="https://goreportcard.com/badge/github.com/sensepost/gowitness" alt="Go Report Card" height="18"></a>
</p>
<br>

## introduction

`gowitness` is a website screenshot utility written in Golang, that uses Chrome Headless to generate screenshots of web interfaces using the command line. Both Linux and macOS is supported, with Windows support 'partially working'.

Inspiration for `gowitness` comes from [Eyewitness](https://github.com/ChrisTruncer/EyeWitness). If you are looking for something with lots of extra features, be sure to check it out along with these [other](https://github.com/afxdub/http-screenshot-html) [projects](https://github.com/breenmachine/httpscreenshot).

## installation

All you would need is an installation of the latest Google Chrome or Chromium and `gowitness` itself. `gowitness` can be downloaded using `go get -u github.com/sensepost/gowitness` or using the
binaries available for download from the [releases](https://github.com/sensepost/gowitness/releases) page.

## build from source

To build `gowitness` from source, follow the following steps:

* Ensure that you have at least golang version 1.11.
* Clone this repository and `cd` into it.
* Run `go build` to get the `gowitness` binary for the current machine.
* Or, `make` to build for all targets. Binaries will be in the `build/` diretory.

## usage examples

### screenshot a single website

`$ gowitness single --url=https://www.google.com/`

This should result in a file being created called: `https-www.google.com.png`

### screenshot a cidr

`$ gowitness scan --cidr 192.168.0.0/24 --threads 20`

This should result in many `.png` images in the current directory when complete. This can would also use `20` threads and not the default of `4`.

### generate a report

`$ gowitness generate`

This should result in an `report.html` file with a screenshot report.

`$ gowitness generate --sort`

This should result in an `report.html` file with a screenshot report where screenshots are sorted using perception hashing.

## license

gowitness is licensed under a [Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International License](http://creativecommons.org/licenses/by-nc-sa/4.0/) Permissions beyond the scope of this license may be available at http://sensepost.com/contact/.
