package restore

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

type RestoreHandler interface {
	RestoreFromSpace(opt *options.SpaceRestoreOption)
	RestoreFromS3(opt *options.S3RestoreOption)
	RestoreFromCos(opt *options.CosRestoreOption)
	RestoreFromFs(opt *options.FilesystemRestoreOption)
}

type RestoreService struct {
	baseDir string
}

func NewRestoreService(baseDir string) RestoreHandler {
	baseDir = util.GetBaseDir(baseDir, common.DefaultBaseDir)

	var restoreService = &RestoreService{
		baseDir: baseDir,
	}
	restoreService.initLogger()
	return restoreService
}

func (r *RestoreService) initLogger() {
	var jsonLogDir = path.Join(r.baseDir, "logs")
	var consoleLogDir = path.Join(r.baseDir, "logs", "backups", "restore.log")

	logger.InitLog(jsonLogDir, consoleLogDir, true)
}

func (r *RestoreService) RestoreFromSpace(opt *options.SpaceRestoreOption) {
	password, err := r.enterPassword()
	if err != nil {
		panic(err)
	}

	var spaceSvc = &space.Space{
		RepoName:       opt.RepoName,
		SnapshotId:     opt.SnapshotId,
		Path:           opt.Path,
		OlaresId:       opt.OlaresId,
		CloudApiMirror: opt.CloudApiMirror,
		BaseDir:        r.baseDir,
		Password:       password,
	}

	if err := spaceSvc.Restore(); err != nil {
		logger.Errorf("Restore from Space error: %v", err)
	}
}

func (r *RestoreService) RestoreFromS3(opt *options.S3RestoreOption) {
	password, err := r.enterPassword()
	if err != nil {
		panic(err)
	}

	var s3Svc = &s3.S3{
		RepoName:        opt.RepoName,
		SnapshotId:      opt.SnapshotId,
		Endpoint:        opt.Endpoint,
		AccessKey:       opt.AccessKey,
		SecretAccessKey: opt.SecretAccessKey,
		Path:            opt.Path,
		Password:        password,
	}

	if err := s3Svc.Restore(); err != nil {
		logger.Errorf("Restore from S3 error: %v", err)
	}
}

func (r *RestoreService) RestoreFromCos(opt *options.CosRestoreOption) {
	password, err := r.enterPassword()
	if err != nil {
		panic(err)
	}

	var cosSvc = &cos.Cos{
		RepoName:        opt.RepoName,
		SnapshotId:      opt.SnapshotId,
		Endpoint:        opt.Endpoint,
		AccessKey:       opt.AccessKey,
		SecretAccessKey: opt.SecretAccessKey,
		Path:            opt.Path,
		Password:        password,
	}

	if err := cosSvc.Restore(); err != nil {
		logger.Errorf("Restore from Tencent COS error: %v", err)
	}
}

func (r *RestoreService) RestoreFromFs(opt *options.FilesystemRestoreOption) {
	password, err := r.enterPassword()
	if err != nil {
		panic(err)
	}

	var filesystemSvc = &filesystem.Filesystem{
		RepoName:   opt.RepoName,
		SnapshotId: opt.SnapshotId,
		Endpoint:   opt.Endpoint,
		Path:       opt.Path,
		Password:   password,
	}

	if err := filesystemSvc.Restore(); err != nil {
		logger.Errorf("Restore from FileSystem error: %v", err)
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
