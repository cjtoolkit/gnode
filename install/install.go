package install

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
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

const cachePath = "/gnode"

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

	b, err := loadFromCache(data)
	if err != nil {
		b, err = downloadAndValidate(data)
		if err != nil {
			cleanSdk(sdkpath)
			log.Fatal(err)
		}
	}

	r := bytes.NewReader(b)

	if data.Ext() == "zip" {
		err = zipInstall(r, int64(len(b)))
	} else {
		err = tarInstall(r)
	}
	if err != nil {
		cleanSdk(sdkpath)
		log.Fatal(err)
	}

	for _, module := range data.Modules {
		err = installModule(binPath, module)
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

func zipInstall(rr io.ReaderAt, size int64) error {
	r, err := zip.NewReader(rr, size)
	if err != nil {
		return err
	}

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}

		path := f.Name

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

func installModule(binPath string, module model.Module) error {
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

func downloadAndValidate(data model.NodeDist) ([]byte, error) {
	client := &http.Client{}
	client.Timeout = 30 * time.Second

	resFile, err := client.Get(data.DistUrl())
	if err != nil {
		return nil, err
	} else if resFile.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Could not find specific version")
	}

	fileBytes, err := ioutil.ReadAll(resFile.Body)
	if err != nil {
		return nil, err
	}

	resHash, err := client.Get(data.DistSumUrl())
	if err != nil {
		return nil, err
	} else if resFile.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Could not file checksum file")
	}

	hashBytes, err := ioutil.ReadAll(resHash.Body)
	if err != nil {
		return nil, err
	}

	fileHash := func() []byte {
		hash := sha256.New()
		hash.Write(fileBytes)
		return hash.Sum(nil)
	}()

	hashAndFile := []byte(fmt.Sprintf("%x  %s", fileHash, data.FileName()))

	if bytes.Index(hashBytes, hashAndFile) == -1 {
		return nil, fmt.Errorf("Checksum mismatch")
	}

	saveToCache(fileBytes, data)

	return fileBytes, nil
}

func saveToCache(fileBytes []byte, data model.NodeDist) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Print(err)
		return
	}

	cacheDir += filepath.FromSlash(cachePath)
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		err = os.MkdirAll(cacheDir, 0755)
	}

	file, err := os.OpenFile(cacheDir+filepath.FromSlash("/"+data.FileName()), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Print(err)
		return
	}
	_, err = file.Write(fileBytes)
	if err != nil {
		log.Print(err)
		return
	}
	err = file.Chdir()
	if err != nil {
		log.Print(err)
		return
	}
}

func loadFromCache(data model.NodeDist) ([]byte, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}

	cacheFileName := cacheDir + filepath.FromSlash(cachePath+"/"+data.FileName())
	if _, err := os.Stat(cacheFileName); os.IsNotExist(err) {
		return nil, err
	}

	file, err := os.Open(cacheFileName)
	if err != nil {
		return nil, err
	}

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return fileBytes, nil
}
