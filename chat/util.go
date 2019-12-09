package chat

import (
	"os/user"
	"path/filepath"
)

// expand a path in case it starts with ~, source: https://stackoverflow.com/a/43578461
func expandPath(path string) (string, error) {
	if len(path) == 0 || path[0] != '~' {
		return path, nil
	}

	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(usr.HomeDir, path[1:]), nil
}
