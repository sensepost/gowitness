<h1 align="center">
  <br>
  <a href="https://github.com/sensepost/objection">
    <img src="images/gowitness-logo.png" alt="objection"></a>
  <br>
  <br>
</h1>

<h4 align="center">A golang, web screenshot utility using Chrome Headless.</h4>
<br>

## introduction
`gowitness` is a website screenshotting tool written in Golang, that uses Chrome Headless to generate screenshots of web interfaces using the command line. Both Linux and macOS is supported, with Windows support 'partially working'.

Inspiration for `gowitness` comes from [Eyewitness](https://github.com/ChrisTruncer/EyeWitness). If you are looking for something with lots of extra features, be sure to check it out.

## installation
Binaries are available for download from the [releases](https://github.com/sensepost/gowitness/releases) page as part of tagged releases.

To build from source, follow the following steps:

* Ensure you have [dep](https://github.com/golang/dep) installed (`go get -v -u github.com/golang/dep/cmd/dep`)
* Clone this repository to your `$GOPATH`'s `src/` directory so that it is in `sensepost/gowitness`
* Run `dep ensure` to resolve dependencies
* Use the `go` build tools, or run `make` to build the binaries in the `build/` directory

## usage
```
~ » gowitness -h
A commandline web screenshot and information gathering tool.

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
      --chrome-timeout int   Time in seconds to wait for Google Chrome to finish a screenshot (default 90)
      --config string        config file (default is $HOME/.gowitness.yaml)
  -D, --db string            Destination for the gowitness database (default "gowitness.db")
  -d, --destination string   Destination directory for screenshots (default ".")
  -h, --help                 help for gowitness
      --log-format string    specify output (text or json) (default "text")
      --log-level string     one of debug, info, warn, error, or fatal (default "info")
  -R, --resolution string    screenshot resolution (default "1440,900")
  -T, --timeout int          Time in seconds to wait for a HTTP connection (default 3)

Use "gowitness [command] --help" for more information about a command.
```

## license

gowtiness is licensed under a [Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International License](http://creativecommons.org/licenses/by-nc-sa/4.0/) Permissions beyond the scope of this license may be available at http://sensepost.com/contact/.