package utils

import (
	"errors"
	"os"
	"os/exec"
	"os/user"
)

func Chmod(p string) error {
	cmd := exec.Command("chown", "-R", "1000:1000", p)
	return cmd.Run()
}

func GetHomeDir() string {
	user, err := user.Current()
	if err != nil {
		panic(errors.New("get current user failed"))
	}

	return user.HomeDir
}

func CreateDir(path string) error {
	if IsExist(path) == false {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err
		}
		return os.Chown(path, 1000, 1000)
	}
	return nil
}

func IsExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		if os.IsNotExist(err) {
			return false
		}
		return false
	}
	return true
}

func Lookup(command string) (string, error) {
	return exec.LookPath(command)
}

func DeleteFile(path string) error {
	return os.Remove(path)
}
