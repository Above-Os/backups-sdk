package snapshots

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

type SnapshotsHandler interface {
	SpaceSnapshots(opt *options.SpaceSnapshotsOption)
	S3Snapshots(opt *options.S3SnapshotsOption)
	CosSnapshots(opt *options.CosSnapshotsOption)
	FsSnapshots(opt *options.FilesystemSnapshotsOption)
}

type SnapshotsService struct {
	baseDir string
}

func NewSnapshotsService(baseDir string) SnapshotsHandler {
	baseDir = util.GetBaseDir(baseDir, common.DefaultBaseDir)

	var snapshotsService = &SnapshotsService{
		baseDir: baseDir,
	}

	snapshotsService.initLogger()

	return snapshotsService
}

func (s *SnapshotsService) initLogger() {
	var jsonLogDir = path.Join(s.baseDir, "logs")
	var consoleLogDir = path.Join(s.baseDir, "logs", "backups", "snapshots.log")

	logger.InitLog(jsonLogDir, consoleLogDir, true)
}

func (s *SnapshotsService) SpaceSnapshots(opt *options.SpaceSnapshotsOption) {
	password, err := s.enterPassword()
	if err != nil {
		panic(err)
	}

	var spaceSvc = &space.Space{
		RepoName:       opt.RepoName,
		OlaresId:       opt.OlaresId,
		CloudApiMirror: opt.CloudApiMirror,
		BaseDir:        s.baseDir,
		Password:       password,
	}

	if err := spaceSvc.Snapshots(); err != nil {
		logger.Errorf("Get Space Snapshots error: %v", err)
	}
}

func (s *SnapshotsService) S3Snapshots(opt *options.S3SnapshotsOption) {
	password, err := s.enterPassword()
	if err != nil {
		panic(err)
	}

	var s3Svc = &s3.S3{
		RepoName:        opt.RepoName,
		Endpoint:        opt.Endpoint,
		AccessKey:       opt.AccessKey,
		SecretAccessKey: opt.SecretAccessKey,
		Password:        password,
	}

	if err := s3Svc.Snapshots(); err != nil {
		logger.Errorf("Get S3 Snapshots error: %v", err)
	}
}

func (s *SnapshotsService) CosSnapshots(opt *options.CosSnapshotsOption) {
	password, err := s.enterPassword()
	if err != nil {
		panic(err)
	}

	var cosSvc = &cos.Cos{
		RepoName:        opt.RepoName,
		Endpoint:        opt.Endpoint,
		AccessKey:       opt.AccessKey,
		SecretAccessKey: opt.SecretAccessKey,
		Password:        password,
	}

	if err := cosSvc.Snapshots(); err != nil {
		logger.Errorf("Get Tencent COS Snapshots error: %v", err)
	}
}

func (s *SnapshotsService) FsSnapshots(opt *options.FilesystemSnapshotsOption) {
	password, err := s.enterPassword()
	if err != nil {
		panic(err)
	}

	var fsSvc = &filesystem.Filesystem{
		RepoName: opt.RepoName,
		Endpoint: opt.Endpoint,
		Password: password,
	}

	if err := fsSvc.Snapshots(); err != nil {
		logger.Errorf("Get FileSystem Snapshots error: %v", err)
	}
}

func (s *SnapshotsService) enterPassword() (string, error) {
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
