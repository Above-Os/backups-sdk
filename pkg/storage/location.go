package storage

import (
	"context"
	"fmt"

	"bytetrade.io/web3os/backups-sdk/pkg/constants"
	"bytetrade.io/web3os/backups-sdk/pkg/logger"
	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/base"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/model"
	"bytetrade.io/web3os/backups-sdk/pkg/utils"
	"github.com/pkg/errors"
)

type Location interface {
	Backup(ctx context.Context, progressCallback func(percentDone float64)) (backupSummary *restic.SummaryOutput, storageInfo *model.StorageInfo, err error)
	Restore(ctx context.Context, progressCallback func(percentDone float64)) (restoreSummary *restic.RestoreSummaryOutput, err error)
	Snapshots(ctx context.Context) (*restic.SnapshotList, error)
	Stats(ctx context.Context) (*restic.StatsContainer, error)
	Regions() ([]map[string]string, error)

	GetEnv(repository string) *restic.ResticEnvs
	FormatRepository() (storageInfo *model.StorageInfo, err error)
}

var _ base.Interface = &BaseHandler{}

type BaseHandler struct {
	location string
	opts     *restic.ResticOptions
}

func (d *BaseHandler) SetOptions(opts *restic.ResticOptions) {
	d.opts = opts
}

func (d *BaseHandler) Backup(ctx context.Context, progressCallback func(percentDone float64)) (backupSummary *restic.SummaryOutput, err error) {
	var traceId = ctx.Value(constants.TraceId).(string)
	var repoName = d.opts.RepoName
	var tags = d.getTags()

	r, err := restic.NewRestic(ctx, d.opts)
	if err != nil {
		return
	}

	var initResult string
	var initialized bool

	// backupType = constants.FullyBackup

	logger.Infof("initializing repo %s, traceId: %s", repoName, traceId)
	initResult, err = r.Init()

	if err != nil {
		if err.Error() == restic.MESSAGE_REPOSITORY_ALREADY_INITIALIZED {
			initialized = true
		} else {
			return
		}
	}

	// if initialized {
	// 	getFullySnapshot, _ := r.GetSnapshots([]string{"type=" + constants.FullyBackup})
	// 	if getFullySnapshot != nil && getFullySnapshot.Len() > 0 {
	// 		backupType = constants.IncrementalBackup
	// 	}
	// }

	if initialized {
		logger.Infof("repo %s already initialized, traceId: %s, repairing index", repoName, traceId)
		if err = r.Repair(); err != nil {
			logger.Errorf("repo %s repair error: %v", repoName, err)
			return
		}
	} else {
		logger.Infof("repo %s initialized, traceId: %s\n\n%s", repoName, traceId, initResult)
	}

	logger.Infof("preparing to start repo %s backup, traceId: %s", repoName, traceId)

	var progressChan = make(chan float64, 100)
	defer close(progressChan)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case progress, ok := <-progressChan:
				if !ok {
					return
				}
				progressCallback(progress)
			}
		}
	}()

	backupSummary, err = r.Backup(d.opts.Path, d.opts.Files, "", tags, traceId, progressChan)
	if err != nil {
		err = errors.WithStack(err)
		if e := r.Rollback(); e != nil {
			err = errors.Wrap(err, e.Error())
		}
		return
	}

	// var currentBackupType = backupType
	// if backupType == constants.FullyBackup {
	// 	var snapshots *restic.SnapshotList
	// 	shortId := backupSummary.SnapshotID[:8]
	// 	logger.Infof("reset tag, name: %s, snapshot: %s, type: %s", repoName, shortId, backupType)
	// 	snapshots, err = r.GetSnapshots(nil)
	// 	if err == nil && snapshots != nil && snapshots.Len() > 0 {
	// 		firstBackup := snapshots.First()
	// 		if firstBackup.Id != backupSummary.SnapshotID {
	// 			currentBackupType = constants.IncrementalBackup
	// 		}
	// 		var resetTags = []string{
	// 			fmt.Sprintf("repo-name=%s", repoName),
	// 			fmt.Sprintf("type=%s", currentBackupType),
	// 		}
	// 		if err = r.Tag(backupSummary.SnapshotID, resetTags); err != nil {
	// 			logger.Errorf("set tag %s error :%v", shortId, err)
	// 			return
	// 		}
	// 	}
	// }

	logger.Infof("Backup successful, result: %s, traceId: %s", utils.ToJSON(backupSummary), traceId)

	return
}

func (h *BaseHandler) Restore(ctx context.Context, progressCallback func(percentDone float64)) (restoreSummary *restic.RestoreSummaryOutput, err error) {
	var snapshotId = h.opts.SnapshotId
	var path = h.opts.Path
	logger.Debugf("restore env vars: %s, snapshotId: %s", utils.Base64encode([]byte(h.opts.RepoEnvs.String())), snapshotId)

	var re *restic.Restic
	re, err = restic.NewRestic(ctx, h.opts)
	if err != nil {
		return
	}
	var snapshotSummary *restic.Snapshot
	snapshotSummary, err = re.GetSnapshot(snapshotId)
	if err != nil {
		return
	}

	var uploadPath = snapshotSummary.Paths[0]

	for _, tag := range snapshotSummary.Tags {
		if tag == "content-type=files" {
			uploadPath = ""
			break
		}
	}

	logger.Infof("restore spanshot %s detail: %s", snapshotId, utils.ToJSON(snapshotSummary))

	var progressChan = make(chan float64, 100)
	defer close(progressChan)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case progress, ok := <-progressChan:
				if !ok {
					return
				}
				progressCallback(progress)
			}
		}
	}()

	restoreSummary, err = re.Restore(snapshotId, uploadPath, path, progressChan)
	if err != nil {
		logger.Errorf("restore %s snapshot %s error: %v", h.opts.RepoName, h.opts.SnapshotId, err)
		return
	}

	logger.Infof("Restore successful, name: %s, result: %s", h.opts.RepoName, utils.ToJSON(restoreSummary))

	return
}

func (h *BaseHandler) Snapshots(ctx context.Context) (*restic.SnapshotList, error) {
	logger.Debugf("snapshots env vars: %s", utils.Base64encode([]byte(h.opts.RepoEnvs.String())))

	r, err := restic.NewRestic(ctx, h.opts)
	if err != nil {
		return nil, err
	}

	snapshots, err := r.GetSnapshots(nil)
	if err != nil {
		return nil, err
	}
	if h.opts.Operator == constants.StorageOperatorCli {
		snapshots.PrintTable()
	}

	return snapshots, nil
}

func (h *BaseHandler) Stats(ctx context.Context) (*restic.StatsContainer, error) {
	logger.Debugf("stats env vars: %s", utils.Base64encode([]byte(h.opts.RepoEnvs.String())))

	r, err := restic.NewRestic(ctx, h.opts)
	if err != nil {
		return nil, err
	}

	stats, err := r.Stats()
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func (h *BaseHandler) getTags() []string {
	var tags = []string{
		fmt.Sprintf("repo-id=%s", h.opts.RepoId),
		fmt.Sprintf("repo-name=%s", h.opts.RepoName),
	}

	if h.opts.Operator != "" {
		tags = append(tags, fmt.Sprintf("operator=%s", h.opts.Operator))
	}

	if h.opts.Files != nil && len(h.opts.Files) > 0 {
		tags = append(tags, "content-type=files")
	} else {
		tags = append(tags, "content-type=dirs")
	}

	return tags
}
