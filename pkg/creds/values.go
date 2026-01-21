package creds

// Credentials are known credential patterns
var Credentials = []*Credential{
	{
		Name: "Integrated Dell Remote Access Controller (iDRAC)",
		Patterns: []string{
			"var thisIDRACText;",
			"thisIDRACText = _jsonData['log_thisDRAC']",
		},
		Credentials: []string{
			"root/calvin",
			"root/<random password>",
		},
		References: []string{
			"https://www.dell.com/support/kbdoc/en-us/000133536/dell-poweredge-what-is-the-default-username-and-password-for-idrac",
		},
	},
	{
		Name: "PRTG Network Monitor",
		Patterns: []string{
			"<link id=\"prtgfavicon\" ",
			"<title>Welcome | PRTG Network Monitor",
			"'appName':'PRTG Network Monitor ",
			"alt=\"The PRTG Network Monitor logo\"",
		},
		Credentials: []string{
			"prtgadmin/prtgadmin",
		},
		References: []string{
			"https://www.paessler.com/manuals/prtg/login",
		},
	},
}
