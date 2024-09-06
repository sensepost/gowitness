package version

var (
	version = "3.0.0-dev"

	gitHash    = ""
	goBuildEnv = ""
)

// Get the current version
func Get() (string, string, string) {

	if gitHash == "" {
		gitHash = "dev"
	}

	if goBuildEnv == "" {
		goBuildEnv = "dev"
	}

	return version, gitHash, goBuildEnv
}
