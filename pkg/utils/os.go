package utils

import (
	"github.com/shirou/gopsutil/v4/host"
)

type Os struct {
	PlatformFamily string
	Type           string
	Arch           string
}

func GetSystem() *Os {
	hostInfo, err := host.Info()
	if err != nil {
		panic(err)
	}

	var platformFamily = formatOsPlatformFamily(hostInfo.Platform, hostInfo.PlatformFamily)
	var arch = formatArch(hostInfo.KernelArch)

	return &Os{
		PlatformFamily: platformFamily,
		Type:           hostInfo.OS,
		Arch:           arch,
	}
}

func formatOsPlatformFamily(osPlatform, osPlatformFamily string) string {
	if osPlatform == "darwin" {
		return "darwin"
	}

	if osPlatform == "raspbian" {
		return osPlatformFamily
	}

	return osPlatform
}

func formatArch(arch string) string {
	switch arch {
	case "aarch64", "armv7l", "arm64", "arm":
		return "arm64"
	case "x86_64", "amd64":
		fallthrough
	case "ppc64le":
		fallthrough
	case "s390x":
		return "amd64"
	default:
		return ""
	}
}
