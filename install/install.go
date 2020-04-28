package install

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/cjtoolkit/gnode/model"
)

func Install(sdkpath, binPath string, data model.NodeDist) {
	err := os.MkdirAll(sdkpath, 0755)
	if err != nil {
		log.Fatal(err)
	}

	curPath, err := os.Getwd()
	if err != nil {
		cleanSdk(sdkpath)
		log.Fatal(err)
	}
	defer os.Chdir(curPath)

	err = os.Chdir(sdkpath)
	if err != nil {
		cleanSdk(sdkpath)
		log.Fatal(err)
	}

	client := &http.Client{}
	client.Timeout = 30 * time.Second
	res, err := client.Get(data.DistUrl())
	if err != nil {
		cleanSdk(sdkpath)
		log.Fatal(err)
	}

	if data.Ext() == "zip" {
		err = zipInstall(res.Body)
	} else {
		err = tarInstall(res.Body)
	}
	if err != nil {
		cleanSdk(sdkpath)
		log.Fatal(err)
	}

	for _, module := range data.Modules {
		err = installModule(sdkpath, binPath, data, module)
		if err != nil {
			cleanSdk(sdkpath)
			log.Fatal(err)
		}
	}
}

func tarInstall(file io.Reader) error {
	r, err := gzip.NewReader(file)
	if err != nil {
		return err
	}

	tr := tar.NewReader(r)
	// Iterate through the files in the archive.
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}
		if err != nil {
			return err
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			// create a directory
			fmt.Println("creating:   " + hdr.Name)
			err = os.Mkdir(hdr.Name, hdr.FileInfo().Mode())
			if err != nil {
				return err
			}
		case tar.TypeReg:
			// write a file
			fmt.Println("extracting: " + hdr.Name)
			w, err := os.OpenFile(hdr.Name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, hdr.FileInfo().Mode())
			if err != nil {
				return err
			}
			_, err = io.Copy(w, tr)
			if err != nil {
				return err
			}
			err = w.Close()
			if err != nil {
				panic(err)
			}
		case tar.TypeSymlink, tar.TypeLink:
			fmt.Println("Creating Symlink: " + hdr.Name)
			err = os.Symlink(hdr.Linkname, hdr.Name)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func zipInstall(file io.Reader) error {
	b, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	r, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return err
	}

	dest, err := os.Getwd()
	if err != nil {
		return err
	}

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}

		path := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			// create a directory
			fmt.Println("creating:   " + path)
			err = os.Mkdir(path, f.Mode())
			if err != nil {
				return err
			}
		} else {
			// write a file
			fmt.Println("extracting: " + path)
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
			err = f.Close()
			if err != nil {
				return err
			}
		}
		err = rc.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func installModule(sdkPath, binPath string, data model.NodeDist, module model.Module) error {
	cmd := exec.Command(filepath.FromSlash(binPath+"/npm"), "install", "-g", module.String())
	cmd.Env = append(os.Environ(), "PATH="+binPath+fmt.Sprintf("%c", os.PathListSeparator)+os.Getenv("PATH"))
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func cleanSdk(sdkPath string) {
	err := os.RemoveAll(sdkPath)
	if err != nil {
		log.Fatal(err)
	}
}
