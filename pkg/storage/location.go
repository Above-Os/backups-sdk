package storage

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"olares.com/backups-sdk/pkg/constants"
	"olares.com/backups-sdk/pkg/logger"
	"olares.com/backups-sdk/pkg/restic"
	"olares.com/backups-sdk/pkg/storage/base"
	"olares.com/backups-sdk/pkg/storage/model"
	"olares.com/backups-sdk/pkg/storage/util"
	"olares.com/backups-sdk/pkg/utils"
)

type Location interface {
	Backup(ctx context.Context, dryRun bool, progressCallback func(percentDone float64)) (backupSummary *restic.SummaryOutput, storageInfo *model.StorageInfo, err error)
	Restore(ctx context.Context, progressCallback func(percentDone float64)) (map[string]*restic.RestoreSummaryOutput, string, uint64, error)
	GetSnapshot(ctx context.Context, snapshotId string) (*restic.SnapshotList, error)
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

func (d *BaseHandler) Backup(ctx context.Context, dryRun bool, progressCallback func(percentDone float64)) (backupSummary *restic.SummaryOutput, err error) {
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
			logger.Errorf("initializing repo %s, traceId: %s, error: %v", repoName, traceId, err)
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

	backupSummary, err = r.Backup(d.opts.Path, d.opts.Files, "", tags, traceId, dryRun, progressChan)
	if err != nil {
		err = errors.WithStack(err)
		if e := r.Rollback(); e != nil {
			logger.Errorf("rollbackup error: %v, traceId: %s", e, traceId)
		}
		// if e := r.Rollback(); e != nil {
		// 	err = errors.Wrap(err, e.Error())
		// }
		return
	}

	restoreSize, _ := r.StatsMode("restore-size")
	if restoreSize != nil {
		backupSummary.RestoreSize = restoreSize.TotalSize
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

func (h *BaseHandler) Restore(ctx context.Context, progressCallback func(percentDone float64)) (map[string]*restic.RestoreSummaryOutput, string, uint64, error) {
	var snapshotId = h.opts.SnapshotId
	var restoreTargetPath = h.opts.Path
	var restoreSummarys = make(map[string]*restic.RestoreSummaryOutput)
	var metadata string
	var totalBytes, totalBytesTmp uint64
	var err error

	logger.Debugf("restore env vars: %s, snapshotId: %s", utils.Base64encode([]byte(h.opts.RepoEnvs.String())), snapshotId)

	var re *restic.Restic
	re, err = restic.NewRestic(ctx, h.opts)
	if err != nil {
		return nil, metadata, totalBytes, err
	}
	var snapshotSummary *restic.Snapshot
	snapshotSummary, err = re.GetSnapshot(snapshotId)
	if err != nil {
		return nil, metadata, totalBytes, err
	}

	logger.Infof("restore spanshot: %s, paths: %d, tags: %v, summary %s", snapshotSummary.Id, len(snapshotSummary.Paths), snapshotSummary.Tags, utils.ToJSON(snapshotSummary.Summary))

	var uploadPaths []string
	var backupMetadata = util.GetMetadata(snapshotSummary.Tags)
	var backupType = util.GetBackupType(snapshotSummary.Tags)

	logger.Infof("restore spanshot: %s, backupType: %s, paths: %d, tags: %v, summary %s", snapshotSummary.Id, backupType, len(snapshotSummary.Paths), snapshotSummary.Tags, utils.ToJSON(snapshotSummary.Summary))

	uploadPaths, _ = util.GetFilesPrefixPath(snapshotSummary.Tags)
	if uploadPaths == nil || len(uploadPaths) == 0 {
		uploadPaths = append(uploadPaths, snapshotSummary.Paths[0])
	}

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

	for phase, uploadPath := range uploadPaths {
		var rs *restic.RestoreSummaryOutput
		var backupTrimPath, targetPath = util.GetRestoreTargetPath(backupType, restoreTargetPath, uploadPath)
		if err = util.Chmod(targetPath); err != nil {
			err = fmt.Errorf("restore %s snapshot %s, backupType: %s, subfolder: %s, create target directory error: %v", h.opts.RepoName, h.opts.SnapshotId, backupType, uploadPath, err)
			break
		}
		rs, err = re.Restore(phase, len(uploadPaths), snapshotId, backupTrimPath, targetPath, progressChan)
		if err != nil {
			logger.Errorf("restore %s snapshot %s, backupType: %s, subfolder: %s, error: %v", h.opts.RepoName, h.opts.SnapshotId, backupType, uploadPath, err)
			break
		}
		if rs != nil {
			restoreSummarys[uploadPath] = rs
			totalBytesTmp += rs.TotalBytes
		}
	}

	if err != nil {
		return nil, metadata, totalBytes, err
	}

	metadata = backupMetadata
	totalBytes = totalBytesTmp

	// restoreSummary, err = re.Restore(snapshotId, uploadPath, path, progressChan)
	// if err != nil {
	// 	logger.Errorf("restore %s snapshot %s error: %v", h.opts.RepoName, h.opts.SnapshotId, err)
	// 	return
	// }

	logger.Infof("Restore successful, name: %s, result: %s", h.opts.RepoName, utils.ToJSON(restoreSummarys))

	return restoreSummarys, metadata, totalBytes, nil
}

func (h *BaseHandler) GetSnapshot(ctx context.Context, snapshotId string) (*restic.SnapshotList, error) {
	logger.Debugf("snapshot env vars: %s", utils.Base64encode([]byte(h.opts.RepoEnvs.String())))

	r, err := restic.NewRestic(ctx, h.opts)
	if err != nil {
		return nil, err
	}

	snapshots, err := r.GetSnapshot(snapshotId)
	if err != nil {
		return nil, err
	}

	var list restic.SnapshotList
	list = append(list, snapshots)

	return &list, nil
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
		fmt.Sprintf("repo-name=%s", utils.Base64encode([]byte(h.opts.RepoName))),
		fmt.Sprintf("backup-type=%s", h.opts.BackupType),
	}

	if h.opts.BackupType == constants.BackupTypeApp {
		tags = append(tags, fmt.Sprintf("backup-app-type-name=%s", utils.Base64encode([]byte(h.opts.BackupAppTypeName))))
	}

	if h.opts.BackupType == constants.BackupTypeFile {
		tags = append(tags, fmt.Sprintf("backup-path=%s", utils.Base64encode([]byte(h.opts.BackupFileTypeSourcePath))))
	}

	if h.opts.Operator != "" {
		tags = append(tags, fmt.Sprintf("operator=%s", h.opts.Operator))
	}

	if h.opts.RepoId != "" {
		tags = append(tags, fmt.Sprintf("repo-id=%s", h.opts.RepoId))
	}

	if h.opts.FilesPrefixPath != "" {
		tags = append(tags, fmt.Sprintf("files-prefix-path=%s", utils.Base64encode([]byte(h.opts.FilesPrefixPath))))
	}

	if h.opts.Metadata != "" {
		tags = append(tags, fmt.Sprintf("metadata=%s", utils.Base64encode([]byte(h.opts.Metadata))))
	}

	return tags
}
