package backup

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

type BackupHandler interface {
	BackupToSpace(opt *options.SpaceOption)
	BackupToS3(opt *options.S3Option)
	BackupToCos(opt *options.CosOption)
	BackupToFilesystem(opt *options.FilesystemOption)
}

type BackupService struct {
	baseDir string
}

func NewBackupService(baseDir string) BackupHandler {
	baseDir = util.GetBaseDir(baseDir, common.DefaultBaseDir)

	var backupService = &BackupService{
		baseDir: baseDir,
	}
	backupService.initLogger()

	return backupService
}

func (b *BackupService) initLogger() {
	var jsonLogDir = path.Join(b.baseDir, "logs")
	var consoleLogDir = path.Join(b.baseDir, "logs", "backups", "backup.log")

	logger.InitLog(jsonLogDir, consoleLogDir, true)
}

func (b *BackupService) BackupToSpace(opt *options.SpaceOption) {
	password, err := b.enterPassword()
	if err != nil {
		panic(err)
	}

	var spaceSvc = &space.Space{
		RepoName:       opt.RepoName,
		Path:           opt.Path,
		OlaresId:       opt.OlaresId,
		CloudApiMirror: opt.CloudApiMirror,
		BaseDir:        b.baseDir,
		Password:       password,
	}

	if err := spaceSvc.Backup(); err != nil {
		logger.Errorf("Backup to Space error: %v", err)
	}
}

func (b *BackupService) BackupToS3(opt *options.S3Option) {
	password, err := b.enterPassword()
	if err != nil {
		panic(err)
	}

	var s3Svc = &s3.S3{
		RepoName:        opt.RepoName,
		Endpoint:        opt.Endpoint,
		AccessKey:       opt.AccessKey,
		SecretAccessKey: opt.SecretAccessKey,
		Path:            opt.Path,
		Password:        password,
	}

	if err := s3Svc.Backup(); err != nil {
		logger.Errorf("Backup to S3 error: %v", err)
	}
}

func (b *BackupService) BackupToCos(opt *options.CosOption) {
	password, err := b.enterPassword()
	if err != nil {
		panic(err)
	}

	var cosSvc = &cos.Cos{
		RepoName:        opt.RepoName,
		Endpoint:        opt.Endpoint,
		AccessKey:       opt.AccessKey,
		SecretAccessKey: opt.SecretAccessKey,
		Path:            opt.Path,
		Password:        password,
	}

	if err := cosSvc.Backup(); err != nil {
		logger.Errorf("Backup to Tencent COS error: %v", err)
	}
}

func (b *BackupService) BackupToFilesystem(opt *options.FilesystemOption) {
	password, err := b.enterPassword()
	if err != nil {
		panic(err)
	}

	var filesystemSvc = &filesystem.Filesystem{
		RepoName: opt.RepoName,
		Endpoint: opt.Endpoint,
		Path:     opt.Path,
		Password: password,
	}

	if err := filesystemSvc.Backup(); err != nil {
		logger.Errorf("Backup to FileSystem error: %v", err)
	}
}

func (b *BackupService) enterPassword() (string, error) {
	fmt.Println("\nPlease create a password for this backup. This password will be required to restore your data in the future. The system will NOT save or store this password, so make sure to remember it. If you lose or forget this password, you will not be able to recover your backup.")

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
