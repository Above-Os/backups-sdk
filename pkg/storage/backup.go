package storage

import (
	"context"
	"fmt"
	"time"

	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/util"
	"bytetrade.io/web3os/backups-sdk/pkg/util/logger"
	"github.com/pkg/errors"
)

type BackupProvider interface {
	Backup() error
}

type Backup struct {
	option BackupOption
	client *StorageClient
}

type BackupOption struct {
	RepoName        string
	OlaresId        string
	StorageLocation string
	Endpoint        string
	AccessKey       string
	SecretAccessKey string
	UploadPath      string
	CloudApiMirror  string
	LimitUploadRate string
	BaseDir         string
	Version         string
}

func (u *Backup) Backup(opt BackupOption) error {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	u.option = opt

	var storageClient = &StorageClient{
		RepoName:        u.option.RepoName,
		OlaresId:        u.option.OlaresId,
		StorageLocation: u.option.StorageLocation,
		Endpoint:        u.option.Endpoint,
		AccessKey:       u.option.AccessKey,
		SecretAccessKey: u.option.SecretAccessKey,
		UploadPath:      u.option.UploadPath,
		CloudApiMirror:  u.option.CloudApiMirror,
		LimitUploadRate: u.option.LimitUploadRate,
		BaseDir:         u.option.BaseDir,
		Version:         u.option.Version,
	}

	u.client = storageClient

	var (
		err     error
		exitCh  = make(chan *StorageResponse)
		summary *restic.SummaryOutput
	)

	go u.run(ctx, exitCh)

	select {
	case e, ok := <-exitCh:
		if ok && e.Error != nil {
			err = e.Error
		}
		summary = e.Summary
	case <-ctx.Done():
		err = errors.Errorf("backup %q osdata timed out in 2 hour", u.option.RepoName)
	}

	if err != nil {
		return err
	}

	if summary != nil {
		logger.Infof("upload successful, data: %s", util.ToJSON(summary))
	}

	return nil
}

func (u *Backup) run(ctx context.Context, exitCh chan<- *StorageResponse) {
	var olaresSpace = &OlaresSpace{
		RepoName:           u.client.RepoName,
		OlaresId:           u.client.OlaresId,
		StorageLocation:    u.client.StorageLocation,
		Endpoint:           u.client.Endpoint,
		AccessKey:          u.client.AccessKey,
		SecretAccessKey:    u.client.SecretAccessKey,
		Path:               u.client.UploadPath,
		CloudApiMirror:     u.client.CloudApiMirror,
		BackupsOperate:     OperateBackup,
		OlaresSpaceSession: new(OlaresSpaceSession),
		BaseDir:            u.client.BaseDir,
		Version:            u.client.Version,
	}

	if err := olaresSpace.GetAccount(); err != nil {
		exitCh <- &StorageResponse{Error: fmt.Errorf("get account error: %v", err)}
		return
	}

	if err := olaresSpace.EnterPassword(); err != nil {
		exitCh <- &StorageResponse{Error: err}
		return
	}

	var summary *restic.SummaryOutput

	if err := olaresSpace.GetToken(); err != nil { // backup
		exitCh <- &StorageResponse{Error: err}
		return
	}

	for {
		olaresSpace.SetRepoUrl() // backup
		olaresSpace.SetEnv()

		logger.Debugf("get token, data: %s", util.Base64encode([]byte(util.ToJSON(olaresSpace))))

		r, err := restic.NewRestic(ctx, u.client.RepoName, u.client.OlaresId, olaresSpace.GetEnv(), &restic.Option{LimitUploadRate: u.client.LimitUploadRate})
		if err != nil {
			exitCh <- &StorageResponse{Error: err}
			return
		}

		var firstInit = true
		_, err = r.Init()
		if err != nil {
			logger.Debugf("restic init message: %s", err.Error())
			if err.Error() == restic.ERROR_MESSAGE_TOKEN_EXPIRED.Error() {
				logger.Infof("olares space token expired, refresh")
				if err := olaresSpace.GetToken(); err != nil { // backup
					exitCh <- &StorageResponse{Error: fmt.Errorf("get token error: %v", err)}
					return
				}
				time.Sleep(2 * time.Second)
				continue
			} else if err.Error() == restic.ERROR_MESSAGE_ALREADY_INITIALIZED.Error() {
				logger.Infof("restic init skip")
				firstInit = false
			} else {
				exitCh <- &StorageResponse{Error: err}
				return
			}
		}

		if !firstInit {
			logger.Infof("restic repair index, please wait...")
			if err := r.Repair(); err != nil {
				exitCh <- &StorageResponse{Error: err}
				return
			}
		}

		logger.Infof("preparing to start backup, repo: %s", olaresSpace.OlaresSpaceSession.ResticRepo)
		summary, err = r.Backup(u.client.RepoName, u.client.UploadPath, "")
		if err != nil {
			switch err.Error() {
			case restic.ERROR_MESSAGE_TOKEN_EXPIRED.Error():
				logger.Infof("olares space token expired, refresh")
				if err := olaresSpace.GetToken(); err != nil { // backup
					exitCh <- &StorageResponse{Error: fmt.Errorf("get token error: %v", err)}
					return
				}
				r.NewContext()
				continue
			default:
				exitCh <- &StorageResponse{Error: err}
				return
			}
		}
		break
	}

	exitCh <- &StorageResponse{Summary: summary}
}
