package files

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"bytetrade.io/web3os/backups-sdk/pkg/util"
)

const (
	DEFAULT_DOWNLOAD_DOMAIN = "https://dc3p1870nn3cj.cloudfront.net"
	RESTIC_VERSION          = "0.17.3"
)

// todo darwin
func Download(downloadCdnUrl string) {
	if err := installBz2(); err != nil {
		panic(err)
	}

	if err := download(downloadCdnUrl); err != nil {
		panic(err)
	}
}

func installBz2() error {
	_, err := util.GetCommand("bzip2")
	if err == nil {
		return nil
	}

	if err := exec.Command("apt-get", "install", "bzip2").Run(); err != nil {
		return err
	}

	return nil
}

func download(downloadCdnUrl string) error {
	if getArch() == "" {
		return errors.New("arch not supported")
	}

	if getOsType() == "" {
		return errors.New("os type not supported")
	}

	filename := getFilename(getOsType())

	url := getDownloadUrl(downloadCdnUrl, util.MD5(filename), getArch())

	cmd := exec.Command("curl", "-o", fmt.Sprintf("%s.bz2", filename), url)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	} else {
		fmt.Println("download success")
	}

	if err := exec.Command("bzip2", "-d", fmt.Sprintf("%s.bz2", filename)).Run(); err != nil {
		return err
	}

	if err := exec.Command("cp", filename, "/usr/local/bin/restic").Run(); err != nil {
		return err
	}

	if err := exec.Command("chmod", "+x", "/usr/local/bin/restic").Run(); err != nil {
		return err
	}

	return nil
}

func getArch() string {
	arch := runtime.GOARCH
	switch arch {
	case "amd64":
		return "amd64"
	case "arm64", "arm":
		return "arm64"
	default:
		return ""
	}
}

func getOsType() string {
	os := runtime.GOOS
	switch os {
	case "linux":
		return "linux"
	case "darwin":
		return "darwin"
	default:
		return ""
	}
}

func getFilename(osType string) string {
	var filename = fmt.Sprintf("restic-%s-%s", osType, RESTIC_VERSION)
	// var h = util.MD5(filename)
	return filename
}

func getDownloadUrl(downloadCdnUrl string, filename string, arch string) string {
	if downloadCdnUrl != "" {
		downloadCdnUrl = strings.TrimRight(downloadCdnUrl, "/")
	} else {
		downloadCdnUrl = DEFAULT_DOWNLOAD_DOMAIN
	}

	var _arch string
	if arch == "arm64" {
		_arch = "/arm64"
	}

	var url = fmt.Sprintf("%s%s/%s", downloadCdnUrl, _arch, filename)
	return url
}
