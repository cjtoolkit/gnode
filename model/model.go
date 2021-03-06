package model

import (
	"crypto/md5"
	"fmt"
	"runtime"
	"sort"
)

const (
	nodeDistUrl    = "https://nodejs.org/dist/v%[1]s/node-v%[1]s-%[2]s-%[3]s.%[4]s"
	nodeDistSumUrl = "https://nodejs.org/dist/v%[1]s/SHASUMS256.txt"
	dirName        = "node-v%[1]s-%[2]s-%[3]s"
	hashDirName    = "node-v%[1]s/%[2]s"
)

type NodeDist struct {
	Version string   `json:"version"`
	NoNpm   bool     `json:"no_npm"`
	Modules []Module `json:"modules"`
}

func (d NodeDist) DistUrl() string {
	return fmt.Sprintf(nodeDistUrl, d.Version, os(), arch(), ext())
}

func (d NodeDist) DistSumUrl() string {
	return fmt.Sprintf(nodeDistSumUrl, d.Version)
}

func (d NodeDist) Dir() string {
	return fmt.Sprintf(dirName, d.Version, os(), arch())
}

func (d NodeDist) FileName() string { return d.Dir() + "." + ext() }

func (d NodeDist) Hash() string {
	hash := md5.New()
	hash.Write([]byte(d.Version))
	if d.NoNpm {
		hash.Write([]byte("no-npm"))
	}
	var l []string
	for _, module := range d.Modules {
		l = append(l, module.String())
	}
	sort.Strings(l)
	for _, v := range l {
		hash.Write([]byte(v))
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func (d NodeDist) DirHash() string {
	return fmt.Sprintf(hashDirName, d.Version, d.Hash())
}

func (d NodeDist) Ext() string { return ext() }

type Module struct {
	Package string `json:"package"`
	Version string `json:"version"`
}

func (m Module) String() string {
	if m.Version != "" {
		return m.Package + "@" + m.Version
	}
	return m.Package
}

func os() string {
	switch runtime.GOOS {
	case "windows":
		return "win"
	case "solaris", "illumos":
		return "sunos"
	}

	return runtime.GOOS
}

func arch() string {
	switch runtime.GOARCH {
	case "386":
		return "x86"
	case "amd64":
		return "x64"
	case "arm":
		return "armv7l"
	}

	return runtime.GOARCH
}

func ext() string {
	if runtime.GOOS == "windows" {
		return "zip"
	}

	return "tar.gz"
}
