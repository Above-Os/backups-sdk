package storage

import (
	"bytes"
	"fmt"
	"log"
	"path"
	"syscall"

	"bytetrade.io/web3os/backups-sdk/pkg/common"
	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/util/logger"
	"golang.org/x/term"
)

type Location interface {
	Backup() error
	Restore() error
	Snapshots() error

	GetEnv(repository string) *restic.ResticEnv
	FormatRepository() (repository string, err error)
	GetRepoName() string
	GetPath() string
	GetSnapshotId() string
	GetLimitUploadRate() string
	GetLocation() common.Location
}

func InitLog(baseDir string, operation common.Operation) {
	var jsonLogDir = path.Join(baseDir, common.DefaultLogDir)
	var consoleLogDir = path.Join(baseDir, common.DefaultLogDir, common.DefaultConsoleLogDir, fmt.Sprintf("%s.log", operation))

	logger.InitLog(jsonLogDir, consoleLogDir, true)
}

func InputPasswordWithConfirm(operation common.Operation) (string, error) {
	if operation == common.Backup {
		fmt.Println("\nPlease create a password for this backup. This password will be required to restore your data in the future. The system will NOT save or store this password, so make sure to remember it. If you lose or forget this password, you will not be able to recover your backup.")
	}

	var password []byte
	var confirmed []byte
	_ = password

	for {
		fmt.Print("\nEnter password for repository: ")
		password, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Fatalf("Failed to read password: %v", err)
			return "", err
		}
		password = bytes.TrimSpace(password)
		if len(password) == 0 {
			continue
		}
		confirmed = password
		if operation != common.Backup {
			break
		}
		fmt.Print("\nRe-enter the password to confirm: ")
		confirmed, err = term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Fatalf("Failed to read re-enter password: %v", err)
			return "", err
		}
		if !bytes.Equal(password, confirmed) {
			fmt.Printf("\nPasswords do not match. Please try again.\n")
			continue
		}

		break
	}
	fmt.Printf("\n\n")

	return string(confirmed), nil
}
