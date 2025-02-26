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

type RestoreProvider interface {
	Restore() error
}

type Restore struct {
	option RestoreOption
	client *StorageClient
}

type RestoreOption struct {
	RepoName          string
	SnapshotId        string
	OlaresId          string
	StorageLocation   string
	Endpoint          string
	AccessKey         string
	SecretAccessKey   string
	TargetPath        string
	CloudApiMirror    string
	LimitDownloadRate string
	BaseDir           string
	Version           string
}

func (d *Restore) Restore(opt RestoreOption) error {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	d.option = opt
	var storageClient = &StorageClient{
		RepoName:          d.option.RepoName,
		SnapshotId:        d.option.SnapshotId,
		OlaresId:          d.option.OlaresId,
		StorageLocation:   d.option.StorageLocation,
		Endpoint:          d.option.Endpoint,
		AccessKey:         d.option.AccessKey,
		SecretAccessKey:   d.option.SecretAccessKey,
		TargetPath:        d.option.TargetPath,
		CloudApiMirror:    d.option.CloudApiMirror,
		LimitDownloadRate: d.option.LimitDownloadRate,
		BaseDir:           d.option.BaseDir,
		Version:           d.option.Version,
	}

	d.client = storageClient

	var (
		err     error
		exitCh  = make(chan *StorageResponse)
		summary *restic.RestoreSummaryOutput
	)

	go d.run(ctx, exitCh)

	select {
	case e, ok := <-exitCh:
		if ok && e.Error != nil {
			err = e.Error
		}
		summary = e.RestoreSummary
	case <-ctx.Done():
		err = errors.Errorf("restore %q osdata timed out in 2 hour", d.option.RepoName)
	}

	if err != nil {
		return err
	}

	if summary != nil {
		logger.Infof("download successful, data: %s", util.ToJSON(summary))
	}

	return nil
}

func (d *Restore) run(ctx context.Context, exitCh chan<- *StorageResponse) {
	var olaresSpace = &OlaresSpace{
		RepoName:           d.client.RepoName,
		OlaresId:           d.client.OlaresId,
		StorageLocation:    d.client.StorageLocation,
		Endpoint:           d.client.Endpoint,
		AccessKey:          d.client.AccessKey,
		SecretAccessKey:    d.client.SecretAccessKey,
		Path:               d.client.TargetPath,
		CloudApiMirror:     d.client.CloudApiMirror,
		BackupsOperate:     OperateRestore,
		OlaresSpaceSession: new(OlaresSpaceSession),
		BaseDir:            d.client.BaseDir,
		Version:            d.client.Version,
	}

	if err := olaresSpace.GetAccount(); err != nil {
		exitCh <- &StorageResponse{Error: fmt.Errorf("get account error: %v", err)}
		return
	}

	if err := olaresSpace.EnterPassword(); err != nil {
		exitCh <- &StorageResponse{Error: err}
		return
	}

	var summary *restic.RestoreSummaryOutput

	if err := olaresSpace.GetToken(); err != nil { // restore
		exitCh <- &StorageResponse{Error: fmt.Errorf("get token error: %v", err)}
		return
	}

	for {
		olaresSpace.SetRepoUrl()
		olaresSpace.SetEnv()

		logger.Debugf("get token, data: %s", util.Base64encode([]byte(util.ToJSON(olaresSpace))))

		r, err := restic.NewRestic(ctx, d.client.RepoName, d.client.OlaresId, olaresSpace.GetEnv(), &restic.Option{LimitDownloadRate: d.client.LimitDownloadRate})
		if err != nil {
			exitCh <- &StorageResponse{Error: err}
			return
		}

		snapshotSummary, err := r.GetSnapshot(d.client.SnapshotId)
		if err != nil {
			exitCh <- &StorageResponse{Error: err}
			return
		}
		var uploadPath = snapshotSummary.Paths[0]

		logger.Infof("snapshot %s detail: %s", d.client.SnapshotId, util.ToJSON(snapshotSummary))

		summary, err = r.Restore(d.client.SnapshotId, uploadPath, d.client.TargetPath)
		if err != nil {
			switch err.Error() {
			case restic.ERROR_MESSAGE_TOKEN_EXPIRED.Error():
				logger.Infof("olares space token expired, refresh")
				if err := olaresSpace.GetToken(); err != nil { // restore
					exitCh <- &StorageResponse{Error: fmt.Errorf("get token error: %v", err)}
					return
				}
				r.NewContext()
				time.Sleep(2 * time.Second)
				continue
			default:
				exitCh <- &StorageResponse{Error: err}
				return
			}
		}
		break
	}

	exitCh <- &StorageResponse{RestoreSummary: summary}
}
