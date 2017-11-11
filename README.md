<h1 align="center">
  <br>
  <a href="https://github.com/sensepost/objection">
    <img src="images/gowitness-logo.png" alt="objection"></a>
  <br>
  <br>
</h1>

<h4 align="center">A golang, web screenshotting utility using Chrome Headless.</h4>
<br>

## introduction
`gowitness` is a website screenshotting tool written in Golang, that used Chrome Headless to capture screenshots of web sites using a commandline interface. Both Linux and macOS is supported, with Windows support 'partially working'.

## features
Scan single URL's, CIDR ranges, or URL's scpecified in a file and optionally generate an HTML report of the results.

## installation
Download one of the prebuild binaries found in the releases page, or build from source!

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