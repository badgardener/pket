package core

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

func Install(pack string, callback Callback) {
	callback.Log("Trying to access home dir...")
	home_dir, err := os.UserHomeDir()

	if err != nil {
		callback.Error("Home directory cannot be defined.")
		return
	}

	var suffix string

	switch runtime.GOOS {
	case "windows":
		suffix = "AppData/Roaming/pket"

	case "linux":
		suffix = ".local/share/pket"

	case "darwin":
		suffix = "Library/Application Support/pket"
	}

	app_home_dir := filepath.Join(home_dir, suffix)
	callback.Log("Detected app home dir: " + app_home_dir)

	callback.Log("Checking home dir...")
	file, err := os.Stat(app_home_dir)

	if err == nil {
		if !file.IsDir() {
			callback.Log("Removing FILE in place of app home dir...")
			err = os.Remove(app_home_dir)

			if err != nil {
				callback.Error("Cannot remove FILE in place of app home dir.")
				return
			}

			callback.Log("Removed FILE in place of app home dir.")
			callback.Log("Creating app home dir...")
			err = os.MkdirAll(app_home_dir, 0755)

			if err != nil {
				callback.Error("Cannot create: " + app_home_dir)
				return
			}

			callback.Log("Home DIR ready.")
		} else {
			callback.Log("Home dir exists.")
		}
	} else {
		callback.Log("Creating app home dir...")
		err = os.MkdirAll(app_home_dir, 0755)

		if err != nil {
			callback.Error("Cannot create: " + app_home_dir)
			return
		}

		callback.Log("Home DIR ready.")
	}

	callback.Log("Checking for target file...")
	file, err = os.Stat(pack)

	if err != nil {
		callback.Error("Target file fat does not exists.")
		return
	}

	callback.Log("Target path stat completed.")
	if file.IsDir() {
		callback.Error("Given path is not a file.")
		return
	}

	callback.Log("Target file exists.")
	install_pack(pack, app_home_dir, callback)

}

func install_pack(pack string, app_dir string, call Callback) {
	call.Warn("Not implemented yet...")
}

func GenerateRandomValue(ignore []string) string {
	sortedIgnore := make([]string, len(ignore))
	copy(sortedIgnore, ignore)
	sort.Strings(sortedIgnore)
	combined := strings.Join(sortedIgnore, "")
	hashBytes := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hashBytes[:])
}
