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

All you would need is an installation of the latest Google Chrome or Chromium and `gowitness` itself. Binaries are available for download from the [releases](https://github.com/sensepost/gowitness/releases) page as part of tagged releases.

To build `gowitness` from source, follow the following steps:

* Ensure you have [dep](https://github.com/golang/dep) installed (`go get -v -u github.com/golang/dep/cmd/dep`)
* Clone this repository to your `$GOPATH`'s `src/` directory so that it is in `sensepost/gowitness`
* Run `dep ensure` to resolve dependencies
* Use the `go` build tools, or run `make` to build the binaries in the `build/` directory

## usage

```txt
~ » gowitness -h
A commandline web screenshot and information gathering tool by @leonjza

Usage:
  gowitness [command]

Available Commands:
  file        Screenshot URLs sourced from a file
  generate    Generate an HTML report from a database file
  help        Help about any command
  scan        Scan a CIDR range and take screenshots along the way
  single      Take a screenshot of a single URL
  version     Prints the version of gowitness

Flags:
      --chrome-path string   Full path to the Chrome executable to use. By default, gowitness will search for Google Chrome
      --chrome-timeout int   Time in seconds to wait for Google Chrome to finish a screenshot (default 90)
      --config string        config file (default is $HOME/.gowitness.yaml)
  -D, --db string            Destination for the gowitness database (default "gowitness.db")
  -d, --destination string   Destination directory for screenshots (default ".")
  -h, --help                 help for gowitness
      --log-format string    specify output (text or json) (default "text")
      --log-level string     one of debug, info, warn, error, or fatal (default "info")
  -R, --resolution string    screenshot resolution (default "1440,900")
  -T, --timeout int          Time in seconds to wait for a HTTP connection (default 3)
      --user-agent string    Alernate UserAgent string to use for Google Chrome (default "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.50 Safari/537.36")

Use "gowitness [command] --help" for more information about a command.
```

## license

gowitness is licensed under a [Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International License](http://creativecommons.org/licenses/by-nc-sa/4.0/) Permissions beyond the scope of this license may be available at http://sensepost.com/contact/.
