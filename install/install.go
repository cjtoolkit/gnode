package install

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
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
		log.Fatal(err)
	}
	defer os.Chdir(curPath)

	err = os.Chdir(sdkpath)
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{}
	client.Timeout = 30 * time.Second
	res, err := client.Get(data.DistUrl())
	if err != nil {
		log.Fatal(err)
	}

	if data.Ext() == "zip" {
		zipInstall(res.Body)
	} else {
		tarInstall(res.Body)
	}

	for _, module := range data.Modules {
		installModule(sdkpath, binPath, data, module)
	}
}

func tarInstall(file io.Reader) {
	r, err := gzip.NewReader(file)
	if err != nil {
		log.Fatal(err)
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
			log.Fatal(err)
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			// create a directory
			fmt.Println("creating:   " + hdr.Name)
			err = os.MkdirAll(hdr.Name, 0755)
			if err != nil {
				log.Fatal(err)
			}
		case tar.TypeReg:
			// write a file
			fmt.Println("extracting: " + hdr.Name)
			w, err := os.Create(hdr.Name)
			if err != nil {
				log.Fatal(err)
			}
			_, err = io.Copy(w, tr)
			if err != nil {
				log.Fatal(err)
			}
			w.Chmod(hdr.FileInfo().Mode())
			w.Close()
		case tar.TypeSymlink, tar.TypeLink:
			fmt.Println("Creating Symlink: " + hdr.Name)
			err = os.Symlink(hdr.Linkname, hdr.Name)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func zipInstall(file io.Reader) {
	buf := &bytes.Buffer{}
	_, err := io.Copy(buf, file)
	if err != nil {
		log.Fatal(err)
	}

	r, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	buf.Reset()
	if err != nil {
		log.Fatal(err)
	}

	dest, err := os.Getwd()

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			fmt.Println("creating:   " + path)
			os.MkdirAll(path, f.Mode())
		} else {
			fmt.Println("extracting: " + path)
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				log.Fatal(err)
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				log.Fatal(err)
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func installModule(sdkPath, binPath string, data model.NodeDist, module model.Module) {
	cmd := exec.Command(filepath.FromSlash(binPath+"/npm"), "install", "-g", module.String())
	cmd.Env = append(os.Environ(), "PATH="+binPath+fmt.Sprintf("%c", os.PathListSeparator)+os.Getenv("PATH"))
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Run()
}
