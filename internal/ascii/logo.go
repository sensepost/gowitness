package ascii

// Logo returns the gowitness ascii logo
func Logo() string {
	return `               _ _                   
 ___ ___ _ _ _|_| |_ __ ___ ___ ___ 
| . | . | | | | |  _|  | -_|_ -|_ -|
|_  |___|_____|_|_||_|_|___|___|___|
|___|    v3, with <3 by @leonjza`
}

// LogoHelp returns the logo, with help
func LogoHelp(s string) string {
	return Logo() + "\n\n" + s
}
