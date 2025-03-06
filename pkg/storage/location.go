package storage

import (
	"context"
	"time"

	"bytetrade.io/web3os/backups-sdk/pkg/logger"
	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/base"
	"bytetrade.io/web3os/backups-sdk/pkg/utils"
)

type Location interface {
	Backup() (err error)
	Restore() error
	Snapshots() error
	Regions() error

	GetEnv(repository string) *restic.ResticEnvs
	FormatRepository() (repository string, err error)
}

var _ base.Interface = &BaseHandler{}

type BaseHandler struct {
	location string
	opts     *restic.ResticOptions
}

func (d *BaseHandler) SetOptions(opts *restic.ResticOptions) {
	d.opts = opts
}

func (d *BaseHandler) Backup() (err error) {
	var repoName = d.opts.RepoName
	var path = d.opts.Path

	r, err := restic.NewRestic(context.Background(), d.opts)
	if err != nil {
		return
	}

	var backupSummary *restic.SummaryOutput
	var initResult string
	var initialized bool

	logger.Infof("initializing repo %s", repoName)
	initResult, err = r.Init()

	if err != nil {
		if err.Error() == restic.MESSAGE_REPOSITORY_ALREADY_INITIALIZED {
			initialized = true
		} else {
			return
		}
	}

	if initialized {
		logger.Infof("repo %s already initialized", repoName)
		logger.Infof("repairing repo %s index", repoName)
		if err = r.Repair(); err != nil {
			return
		}
	} else {
		logger.Infof("repo %s initialized\n\n%s", repoName, initResult)
	}

	backupSummary, err = r.Backup(path, "", nil)
	if err != nil {
		return
	}

	logger.Info("Backup successful, result: ", utils.ToJSON(backupSummary))

	return
}

func (h *BaseHandler) Restore() error {
	var snapshotId = h.opts.SnapshotId
	var path = h.opts.Path
	logger.Debugf("restore env vars: %s", utils.Base64encode([]byte(h.opts.RepoEnvs.String())))

	re, err := restic.NewRestic(context.Background(), h.opts)
	if err != nil {
		return err
	}
	snapshotSummary, err := re.GetSnapshot(snapshotId)
	if err != nil {
		return err
	}

	var uploadPath = snapshotSummary.Paths[0]
	logger.Infof("restore spanshot %s detail: %s", snapshotId, utils.ToJSON(snapshotSummary))

	var summary *restic.RestoreSummaryOutput
	summary, err = re.Restore(snapshotId, uploadPath, path)
	if err != nil {
		return err
	}

	if summary != nil {
		logger.Infof("Restore successful, data: %s", utils.ToJSON(summary))
	}

	return nil
}

func (h *BaseHandler) Snapshots() error {
	logger.Debugf("snapshots env vars: %s", utils.Base64encode([]byte(h.opts.RepoEnvs.String())))

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	r, err := restic.NewRestic(ctx, h.opts)
	if err != nil {
		return err
	}

	snapshots, err := r.GetSnapshots(nil)
	if err != nil {
		return err
	}
	snapshots.PrintTable()

	return nil
}
