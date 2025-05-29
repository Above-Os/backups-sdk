package space

import (
	"context"
	"fmt"

	"bytetrade.io/web3os/backups-sdk/pkg/constants"
	"bytetrade.io/web3os/backups-sdk/pkg/logger"
	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/model"
	"bytetrade.io/web3os/backups-sdk/pkg/utils"
	"github.com/pkg/errors"
)

func (s *Space) Backup(ctx context.Context, dryRun bool, progressCallback func(percentDone float64)) (backupSummary *restic.SummaryOutput, storageInfo *model.StorageInfo, err error) {
	if err = s.getStsToken(ctx); err != nil {
		return
	}

	storageInfo, err = s.FormatRepository()
	if err != nil {
		return
	}

	repoSuffix, err := utils.GetSuffix(s.StsToken.Prefix, "-")
	if err != nil {
		return
	}

	var traceId = ctx.Value(constants.TraceId).(string)

	// backupType = constants.FullyBackup

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
				if !dryRun {
					progressCallback(progress)
				}
			}
		}
	}()

	for {
		var initResult string
		var initialized bool

		var envs = s.GetEnv(storageInfo.Url)
		var opts = &restic.ResticOptions{
			RepoId:          s.RepoId,
			RepoName:        s.RepoName,
			RepoSuffix:      repoSuffix,
			CloudName:       s.CloudName,
			RegionId:        s.RegionId,
			Operator:        s.Operator,
			BackupType:      s.BackupType,
			RepoEnvs:        envs,
			LimitUploadRate: s.LimitUploadRate,
		}

		logger.Infof("space backup env vars: %s, traceId: %s", utils.Base64encode([]byte(envs.String())), traceId)

		var r *restic.Restic
		r, err = restic.NewRestic(ctx, opts)
		if err != nil {
			break
		}

		logger.Infof("initializing repo %s, traceId: %s", s.RepoName, traceId)
		initResult, err = r.Init()
		if err != nil {
			if err.Error() == restic.MESSAGE_REPOSITORY_ALREADY_INITIALIZED {
				initialized = true
			} else {
				logger.Errorf("error initializing repo %s, err: %s, traceId: %s", s.RepoName, err.Error(), traceId)
				break
			}
		}

		// if initialized {
		// 	getFullySnapshot, _ := r.GetSnapshots([]string{"type=" + constants.FullyBackup})
		// 	if getFullySnapshot != nil && getFullySnapshot.Len() > 0 {
		// 		backupType = constants.IncrementalBackup
		// 	}
		// }

		if initialized {
			logger.Infof("repo %s already initialized, traceId: %s", s.RepoName, traceId)
			logger.Infof("repairing repo %s index, traceId: %s", s.RepoName, traceId)
			if err = r.Repair(); err != nil {
				break
			}
		} else {
			logger.Infof("repo %s initialized, traceId: %s\n\n%s", s.RepoName, traceId, initResult)
		}

		logger.Infof("preparing to start repo %s backup, traceId: %s", s.RepoName, traceId)

		var tags = s.getTags()
		tags = append(tags, fmt.Sprintf("repo-suffix=%s", repoSuffix))

		backupSummary, err = r.Backup(s.Path, s.Files, "", tags, traceId, dryRun, progressChan)
		if err != nil {
			// switch err.Error() {
			// case restic.ERROR_MESSAGE_BACKUP_CANCELED.Error():
			// 	logger.Infof("backup canceled, stopping..., traceId: %s", traceId)
			// 	return
			// case restic.ERROR_MESSAGE_TOKEN_EXPIRED.Error():
			// 	logger.Infof("space backup upload stopped, sts token expired, refresh and retring..., traceId: %s", traceId)
			// 	if err = s.refreshStsTokens(ctx); err != nil {
			// 		err = fmt.Errorf("space backup upload sts token service refresh-token error: %v, traceId: %s", err, traceId)
			// 		return
			// 	}
			// 	continue
			// default:
			// 	return nil, nil, errors.WithStack(err)
			// }

			if !dryRun {
				if err == restic.ERROR_MESSAGE_TOKEN_EXPIRED {
					if err = s.refreshStsTokens(ctx); err == nil {
						continue
					} else {
						logger.Errorf("space backup upload sts token service refresh-token error: %v, traceId: %s", err, traceId)
						// err = fmt.Errorf("space backup upload sts token service refresh-token error: %v, traceId: %s", err, traceId)
					}
				}

				e := r.Rollback()
				if e != nil {
					err = errors.Wrap(err, e.Error())
				}
			}
			break
		}

		// var currentBackupType = backupType
		// if backupType == constants.FullyBackup {
		// 	shortId := backupSummary.SnapshotID[:8]
		// 	logger.Infof("reset tag, name: %s, snapshot: %s, type: %s", s.RepoName, shortId, backupType)
		// 	snapshots, err := r.GetSnapshots(nil)
		// 	if err == nil && snapshots != nil && snapshots.Len() > 0 {
		// 		firstBackup := snapshots.First()
		// 		if firstBackup.Id != backupSummary.SnapshotID {
		// 			currentBackupType = constants.IncrementalBackup
		// 		}
		// 		var resetTags = []string{
		// 			fmt.Sprintf("repo-name=%s", s.RepoName),
		// 			fmt.Sprintf("type=%s", currentBackupType),
		// 		}
		// 		if err := r.Tag(backupSummary.SnapshotID, resetTags); err != nil {
		// 			logger.Errorf("set tag %s error :%v", shortId, err)
		// 			break
		// 		}
		// 	}
		// }

		logger.Infof("Backup successful, name: %s, result: %s, traceId: %s", s.RepoName, utils.ToJSON(backupSummary), traceId)
		// if err := s.sendBackup(backupSummary, currentBackupType, opts.RepoEnvs.RESTIC_REPOSITORY); err != nil {
		// 	logger.Errorf("send backup to cloud error: %v", err)
		// }
		break
	}

	return
}

// func (s *Space) sendBackup(backupResult *restic.SummaryOutput, backupType string, backupUrl string) error {
// 	cloudApiUrl := s.getCloudApi()
// 	if backupType == constants.FullyBackup {
// 		var backupData = &notification.Backup{
// 			UserId:         s.OlaresDid,
// 			Token:          s.AccessToken,
// 			BackupId:       "",
// 			Name:           s.RepoName,
// 			BackupPath:     s.Path,
// 			BackupLocation: s.CloudName,
// 			Status:         constants.BackupComplete,
// 		}

// 		if err := notification.SendNewBackup(cloudApiUrl, backupData); err != nil {
// 			return err
// 		}
// 	}

// 	var snapshotData = &notification.Snapshot{
// 		UserId:       s.OlaresDid,
// 		BackupId:     "",
// 		SnapshotId:   backupResult.SnapshotID,
// 		Size:         backupResult.TotalBytesProcessed,
// 		Uint:         "byte",
// 		SnapshotTime: time.Now().UnixMilli(),
// 		Status:       constants.BackupComplete,
// 		Type:         backupType,
// 		Url:          backupUrl,
// 		CloudName:    s.CloudName,
// 		RegionId:     s.RegionId,
// 		Bucket:       s.StsToken.Bucket,
// 		Prefix:       s.StsToken.Prefix,
// 		Message:      utils.ToJSON(backupResult),
// 	}

// 	if err := notification.SendNewSnapshot(cloudApiUrl, snapshotData); err != nil {
// 		return err
// 	}
// 	return nil

// }
