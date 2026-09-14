package driver

import "os"

// resolveUserDataDir decides which Chrome user data directory a driver should
// use. When configured is empty a fresh temporary directory is created and
// returned with ephemeral set to true, so the caller knows to remove it when
// the browser closes. When the user pointed us at their own directory it is
// returned as is with ephemeral false, so we never delete a real profile.
func resolveUserDataDir(configured, pattern string) (dir string, ephemeral bool, err error) {
	if configured != "" {
		return configured, false, nil
	}

	dir, err = os.MkdirTemp("", pattern)
	if err != nil {
		return "", false, err
	}

	return dir, true, nil
}
