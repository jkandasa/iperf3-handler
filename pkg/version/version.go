package version

import (
	"fmt"
	"runtime"
	"sync"
)

var (
	gitCommit string
	version   string
	buildDate string

	runOnce  sync.Once
	_version Version
)

// Version holds version data
type Version struct {
	Version   string `json:"version" yaml:"version"`
	GitCommit string `json:"gitCommit" yaml:"gitCommit"`
	BuildDate string `json:"buildDate" yaml:"buildDate"`
	GoVersion string `json:"goVersion" yaml:"goVersion"`
	Compiler  string `json:"compiler" yaml:"compiler"`
	Platform  string `json:"platform" yaml:"platform"`
	Arch      string `json:"arch" yaml:"arch"`
}

// Get returns the Version object
func Get() Version {
	runOnce.Do(func() {
		_version = Version{
			GitCommit: gitCommit,
			Version:   version,
			BuildDate: buildDate,
			GoVersion: runtime.Version(),
			Compiler:  runtime.Compiler,
			Platform:  runtime.GOOS,
			Arch:      runtime.GOARCH,
		}
	})
	return _version
}

// String returns the values as string
func (v Version) String() string {
	return fmt.Sprintf("{version:%s, buildDate:%s, gitCommit:%s, goVersion:%s, compiler:%s, platform:%s, arch:%s}",
		v.Version, v.BuildDate, v.GitCommit, v.GoVersion, v.Compiler, v.Platform, v.Arch)
}
