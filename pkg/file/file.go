package file

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"bytetrade.io/web3os/backups-sdk/pkg/constants"
	"bytetrade.io/web3os/backups-sdk/pkg/utils"
)

const (
	DefaultResticVersion = "0.17.3"
)

func Download(downloadCdnUrl string) error {
	os := utils.GetSystem()
	pkg := getPackageManager(os.PlatformFamily)
	if err := checkAppDeps(pkg); err != nil {
		return fmt.Errorf("Dependency check for curl, bzip2 failed. Please manually install curl or bzip2, error: %v", err)
	}

	if err := download(os.Type, os.Arch, downloadCdnUrl); err != nil {
		return err
	}

	return nil
}

func getPackageManager(platformFamily string) string {
	switch platformFamily {
	case "ubuntu", "debian":
		return "apt-get"
	case "fedora":
		return "dnf"
	case "centos", "rhel":
		return "yum"
	default:
		return "apt-get"
	}
}

func checkApp(pkg string, app string) error {
	if _, err := utils.Lookup(app); err != nil {
		var cmd = exec.Command(pkg, "install", "-y", app)
		fmt.Printf("run %s\n", cmd.String())
		return cmd.Run()
	}
	return nil
}

func checkAppDeps(pkg string) error {
	for _, app := range []string{"curl", "bzip2"} {
		checkApp(pkg, app)
	}

	return nil
}

func download(ostype string, arch string, downloadCdnUrl string) error {
	filename := getFileName(ostype, arch)
	url := getDownloadUrl(downloadCdnUrl, utils.MD5(filename), arch)
	cmd := exec.Command("curl", "-o", fmt.Sprintf("%s.bz2", filename), url)
	fmt.Printf("run: %s\n", cmd.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	} else {
		fmt.Println("download success")
	}

	if err := exec.Command("bzip2", "-d", fmt.Sprintf("%s.bz2", filename)).Run(); err != nil {
		fmt.Printf("bzip2 %s error: %v\n", filename, err)
		return err
	}

	if err := exec.Command("cp", filename, "/usr/local/bin/restic").Run(); err != nil {
		fmt.Printf("cp %s error: %v\n", filename, err)
		return err
	}

	if err := exec.Command("chmod", "+x", "/usr/local/bin/restic").Run(); err != nil {
		fmt.Printf("chmod %s error: %v\n", filename, err)
		return err
	}

	return nil
}

func getFileName(ostype, arch string) string {
	var filename = fmt.Sprintf("restic-%s-%s", ostype, DefaultResticVersion)
	return filename
}

func getDownloadUrl(downloadCdnUrl string, filename string, arch string) string {
	if downloadCdnUrl != "" {
		downloadCdnUrl = strings.TrimRight(downloadCdnUrl, "/")
	} else {
		downloadCdnUrl = constants.DefaultDownloadUrl
	}

	downloadCdnUrl = utils.TrimRight(downloadCdnUrl, "/")
	var archType string
	if arch == "arm64" {
		archType = "/arm64"
	}

	var url = fmt.Sprintf("%s%s/%s", downloadCdnUrl, archType, filename)
	return url
}
