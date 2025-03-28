package space

import (
	"context"
	"fmt"
	"time"

	"bytetrade.io/web3os/backups-sdk/pkg/constants"
	"bytetrade.io/web3os/backups-sdk/pkg/logger"
	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/notification"
	"bytetrade.io/web3os/backups-sdk/pkg/utils"
	"github.com/pkg/errors"
)

func (s *Space) Backup() (err error) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	if err = s.getStsToken(); err != nil {
		return
	}

	var backupId = utils.MD5(fmt.Sprintf("%s_%s", s.RepoName, s.ClusterId))
	var backupType = constants.FullyBackup
	var backupSummary *restic.SummaryOutput

	for {
		var initResult string
		var initialized bool

		var envs = s.GetEnv(s.RepoName)
		var opts = &restic.ResticOptions{
			RepoName:        s.RepoName,
			CloudName:       s.CloudName,
			RegionId:        s.RegionId,
			RepoEnvs:        envs,
			LimitUploadRate: s.LimitUploadRate,
		}

		logger.Debugf("space backup env vars: %s", utils.Base64encode([]byte(envs.String())))

		var r *restic.Restic
		r, err = restic.NewRestic(ctx, opts)
		if err != nil {
			break
		}

		logger.Infof("initializing repo %s", s.RepoName)
		initResult, err = r.Init()
		if err != nil {
			if err.Error() == restic.MESSAGE_REPOSITORY_ALREADY_INITIALIZED {
				initialized = true
			} else {
				break
			}
		}

		if initialized {
			getFullySnapshot, _ := r.GetSnapshots([]string{"type=" + constants.FullyBackup})
			if getFullySnapshot != nil && getFullySnapshot.Len() > 0 {
				backupType = constants.IncrementalBackup
			}
		}

		if initialized {
			logger.Infof("repo %s already initialized", s.RepoName)
			logger.Infof("repairing repo %s index", s.RepoName)
			if err = r.Repair(); err != nil {
				break
			}
		} else {
			logger.Infof("repo %s initialized\n\n%s", s.RepoName, initResult)
		}

		logger.Infof("preparing to start repo %s backup", s.RepoName)

		var tags = []string{
			fmt.Sprintf("repo-name=%s", s.RepoName),
			fmt.Sprintf("backup-id=%s", backupId),
			fmt.Sprintf("type=%s", backupType)}

		backupSummary, err = r.Backup(s.Path, "", tags)
		if err != nil {
			switch err.Error() {
			case restic.ERROR_MESSAGE_TOKEN_EXPIRED.Error():
				logger.Infof("space backup upload stopped, sts token expired, refresh and retring...")
				if err = s.refreshStsTokens(); err != nil {
					err = fmt.Errorf("space backup upload sts token service refresh-token error: %v", err)
					return
				}
				continue
			default:
				return errors.WithStack(err)
			}
		}

		var currentBackupType = backupType
		if backupType == constants.FullyBackup {
			shortId := backupSummary.SnapshotID[:8]
			logger.Infof("reset tag, name: %s, id: %s , snapshot: %s, type: %s", s.RepoName, backupId, shortId, backupType)
			snapshots, err := r.GetSnapshots(nil)
			if err == nil && snapshots != nil && snapshots.Len() > 0 {
				firstBackup := snapshots.First()
				if firstBackup.Id != backupSummary.SnapshotID {
					currentBackupType = constants.IncrementalBackup
				}
				var resetTags = []string{
					fmt.Sprintf("repo-name=%s", s.RepoName),
					fmt.Sprintf("backup-id=%s", backupId),
					fmt.Sprintf("type=%s", currentBackupType),
				}
				if err := r.Tag(backupSummary.SnapshotID, resetTags); err != nil {
					logger.Errorf("set tag %s error :%v", shortId, err)
					break
				}
			}
		}

		logger.Infof("Backup successful, name: %s, type: %s, result: %s", s.RepoName, currentBackupType, utils.ToJSON(backupSummary))
		if err := s.sendBackup(backupSummary, backupId, currentBackupType, opts.RepoEnvs.RESTIC_REPOSITORY); err != nil {
			logger.Errorf("send backup to cloud error: %v", err)
		}
		break
	}

	return
}

func (s *Space) sendBackup(backupResult *restic.SummaryOutput, backupId string, backupType string, backupUrl string) error {
	cloudApiUrl := s.getCloudApi()
	if backupType == constants.FullyBackup {
		var backupData = &notification.Backup{
			UserId:         s.OlaresDid,
			Token:          s.AccessToken,
			BackupId:       backupId,
			Name:           s.RepoName,
			BackupPath:     s.Path,
			BackupLocation: s.CloudName,
			Status:         constants.BackupComplete,
		}

		if err := notification.SendNewBackup(cloudApiUrl, backupData); err != nil {
			return err
		}
	}

	var snapshotData = &notification.Snapshot{
		UserId:       s.OlaresDid,
		BackupId:     backupId,
		SnapshotId:   backupResult.SnapshotID,
		Size:         backupResult.TotalBytesProcessed,
		Uint:         "byte",
		SnapshotTime: time.Now().UnixMilli(),
		Status:       constants.BackupComplete,
		Type:         backupType,
		Url:          backupUrl,
		CloudName:    s.CloudName,
		RegionId:     s.RegionId,
		Bucket:       s.StsToken.Bucket,
		Prefix:       s.StsToken.Prefix,
		Message:      utils.ToJSON(backupResult),
	}

	if err := notification.SendNewSnapshot(cloudApiUrl, snapshotData); err != nil {
		return err
	}
	return nil

}
