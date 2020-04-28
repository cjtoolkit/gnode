package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cjtoolkit/gnode/install"
	"github.com/cjtoolkit/gnode/model"
)

func seekGnodeFile() string {
	curdir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	gnode := filepath.FromSlash("/.gnode")
	lastPath := false
	for {
		if _, err := os.Stat(curdir + gnode); os.IsNotExist(err) {
			curdir = filepath.Dir(curdir)
			if lastPath || curdir == "." {
				log.Fatal("Could not find '.gnode'")
			} else if strings.Trim(curdir, "/") == "" || (runtime.GOOS == "windows" && len(strings.Trim(curdir, "\\")) == 2) {
				lastPath = true
			}
			continue
		}
		break
	}
	return curdir + gnode
}

func main() {
	gnodePath := seekGnodeFile()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	file, err := os.Open(gnodePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	data := model.NodeDist{}
	err = json.NewDecoder(file).Decode(&data)
	if err != nil {
		log.Fatal(err)
	}

	sdkPath := homeDir + filepath.FromSlash("/sdk/"+data.DirHash())

	binPath := sdkPath + filepath.FromSlash("/"+data.Dir())
	if runtime.GOOS != "windows" {
		binPath += filepath.FromSlash("/bin")
	}

	if _, err := os.Stat(sdkPath); os.IsNotExist(err) {
		install.Install(sdkPath, binPath, data)
	}

	if len(os.Args) <= 1 {
		return
	}

	cmd := exec.Command(filepath.FromSlash(binPath+"/"+os.Args[1]), os.Args[2:]...)
	cmd.Env = append(os.Environ(), "PATH="+binPath+fmt.Sprintf("%c", os.PathListSeparator)+os.Getenv("PATH"))
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}
