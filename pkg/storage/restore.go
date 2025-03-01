package storage

import (
	"bytes"
	"fmt"
	"log"
	"path"
	"syscall"

	"bytetrade.io/web3os/backups-sdk/cmd/options"
	"bytetrade.io/web3os/backups-sdk/pkg/common"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/cos"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/filesystem"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/s3"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/space"
	"bytetrade.io/web3os/backups-sdk/pkg/util"
	"bytetrade.io/web3os/backups-sdk/pkg/util/logger"
	"golang.org/x/term"
)

type RestoreOption struct {
	Basedir    string
	Space      *options.SpaceRestoreOption
	S3         *options.S3RestoreOption
	Cos        *options.CosRestoreOption
	Filesystem *options.FilesystemRestoreOption
}

type RestoreService struct {
	baseDir string
	option  *RestoreOption
}

func NewRestoreService(option *RestoreOption) *RestoreService {
	baseDir := util.GetBaseDir(option.Basedir, common.DefaultBaseDir)

	var restoreService = &RestoreService{
		baseDir: baseDir,
		option:  option,
	}
	var jsonLogDir = path.Join(baseDir, "logs")
	var consoleLogDir = path.Join(baseDir, "logs", "backups", "restore.log")

	logger.InitLog(jsonLogDir, consoleLogDir, true)

	return restoreService
}

func (r *RestoreService) Restore() {
	password, err := r.enterPassword()
	if err != nil {
		panic(err)
	}

	var service Location

	if r.option.Space != nil {
		service = &space.Space{
			RepoName:       r.option.Space.RepoName,
			SnapshotId:     r.option.Space.SnapshotId,
			Path:           r.option.Space.Path,
			OlaresId:       r.option.Space.OlaresId,
			CloudApiMirror: r.option.Space.CloudApiMirror,
			BaseDir:        r.baseDir,
			Password:       password,
		}
	} else if r.option.S3 != nil {
		service = &s3.S3{
			RepoName:        r.option.S3.RepoName,
			SnapshotId:      r.option.S3.SnapshotId,
			Endpoint:        r.option.S3.Endpoint,
			AccessKey:       r.option.S3.AccessKey,
			SecretAccessKey: r.option.S3.SecretAccessKey,
			Path:            r.option.S3.Path,
			Password:        password,
		}
	} else if r.option.Cos != nil {
		service = &cos.Cos{
			RepoName:        r.option.Cos.RepoName,
			SnapshotId:      r.option.Cos.SnapshotId,
			Endpoint:        r.option.Cos.Endpoint,
			AccessKey:       r.option.Cos.AccessKey,
			SecretAccessKey: r.option.Cos.SecretAccessKey,
			Path:            r.option.Cos.Path,
			Password:        password,
		}

	} else if r.option.Filesystem != nil {
		service = &filesystem.Filesystem{
			RepoName:   r.option.Filesystem.RepoName,
			SnapshotId: r.option.Filesystem.SnapshotId,
			Endpoint:   r.option.Filesystem.Endpoint,
			Path:       r.option.Filesystem.Path,
			Password:   password,
		}
	} else {
		logger.Fatalf("There is no suitable recovery method.")
	}

	if err := service.Restore(); err != nil {
		logger.Errorf("Restore from Space error: %v", err)
	}
}

func (r *RestoreService) enterPassword() (string, error) {
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
		break
	}
	fmt.Printf("\n\n")

	return string(confirmed), nil
}
