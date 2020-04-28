package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/cjtoolkit/gnode/install"
	"github.com/cjtoolkit/gnode/model"
)

func main() {
	if _, err := os.Stat(".gnode"); os.IsNotExist(err) {
		log.Fatal("'.gnode' file does not exist")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	file, err := os.Open(".gnode")
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
