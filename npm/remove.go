package npm

import (
	"os"
	"path/filepath"
	"runtime"
)

func Remove(binPath string) {
	if runtime.GOOS == "windows" {
		os.Remove(binPath + filepath.FromSlash("/npm"))
		os.Remove(binPath + filepath.FromSlash("/npm.cmd"))
		os.Remove(binPath + filepath.FromSlash("/npx"))
		os.Remove(binPath + filepath.FromSlash("/npx.cmd"))
		os.RemoveAll(binPath + filepath.FromSlash("/node_modules/npm"))
		return
	}

	os.Remove(binPath + filepath.FromSlash("/npm"))
	os.Remove(binPath + filepath.FromSlash("/npx"))
	os.RemoveAll(filepath.Dir(binPath) + filepath.FromSlash("/lib/node_modules/npm"))
}
