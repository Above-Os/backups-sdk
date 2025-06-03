package space

import (
	"context"
	"fmt"

	"olares.com/backups-sdk/pkg/constants"
	"olares.com/backups-sdk/pkg/logger"
	"olares.com/backups-sdk/pkg/restic"
	"olares.com/backups-sdk/pkg/storage/util"
	"olares.com/backups-sdk/pkg/utils"
)

func (s *Space) Restore(ctx context.Context, progressCallback func(percentDone float64)) (map[string]*restic.RestoreSummaryOutput, string, uint64, error) {
	// ctx, cancel := context.WithCancel(context.TODO())
	// defer cancel()

	var err error
	var restoreSummarys = make(map[string]*restic.RestoreSummaryOutput)
	var metadata, backupMetadata string
	var totalBytes, totalBytesTmp uint64

	if err = s.getStsToken(ctx); err != nil {
		return nil, metadata, totalBytes, err
	}

	storageInfo, err := s.FormatRepository()
	if err != nil {
		return nil, metadata, totalBytes, err
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

	var restoreTargetPath = s.Path

	for {
		var envs = s.GetEnv(storageInfo.Url)
		var opts = &restic.ResticOptions{
			RepoId:            s.RepoId,
			RepoName:          s.RepoName,
			CloudName:         s.CloudName,
			RegionId:          s.RegionId,
			RepoEnvs:          envs,
			Path:              s.Path,
			LimitDownloadRate: s.LimitDownloadRate,
		}

		logger.Debugf("space restore env vars: %s", utils.Base64encode([]byte(envs.String())))

		var r *restic.Restic
		r, err = restic.NewRestic(ctx, opts)
		if err != nil {
			break
		}

		var currentSnapshot *restic.Snapshot
		currentSnapshot, err = r.GetSnapshot(s.SnapshotId)
		if err != nil {
			break
		}

		var uploadPaths []string
		var backupType = util.GetBackupType(currentSnapshot.Tags)
		backupMetadata = util.GetMetadata(currentSnapshot.Tags)

		logger.Infof("space restore spanshot: %s, backupType: %s, paths: %d, tags: %v, summary %s", currentSnapshot.Id, backupType, len(currentSnapshot.Paths), currentSnapshot.Tags, utils.ToJSON(currentSnapshot.Summary))

		if backupType == constants.BackupTypeFile {
			uploadPaths = append(uploadPaths, currentSnapshot.Paths[0])
		} else {
			uploadPaths, err = util.GetFilesPrefixPath(currentSnapshot.Tags)
			if err != nil {
				break
			}
		}

		// var backupPath = currentSnapshot.Paths[0]

		// for _, tag := range currentSnapshot.Tags {
		// 	if tag == "content-type=files" {
		// 		backupPath = ""
		// 		break
		// 	}
		// }

		// logger.Infof("space restore spanshot %s detail: %s", s.SnapshotId, utils.ToJSON(currentSnapshot))

		for _, uploadPath := range uploadPaths {
			var rs *restic.RestoreSummaryOutput
			var backupTrimPath, targetPath = util.GetRestoreTargetPath(backupType, restoreTargetPath, uploadPath)
			if err = util.Chmod(targetPath); err != nil {
				err = fmt.Errorf("space restore %s snapshot %s, backupType: %s, subfolder: %s, create target directory error: %v", s.RepoName, s.SnapshotId, backupType, uploadPath, err)
				break
			}
			rs, err = r.Restore(s.SnapshotId, backupTrimPath, targetPath, progressChan)
			if err != nil {
				switch err.Error() {
				case restic.ERROR_MESSAGE_TOKEN_EXPIRED.Error():
					logger.Infof("space restore download stopped, sts token expired, refresh and retring...")
					if err = s.refreshStsTokens(ctx); err != nil {
						err = fmt.Errorf("space restore download sts token service refresh-token error: %v", err)
						break
					}
					continue
				default:
					break
				}
			}
			restoreSummarys[uploadPath] = rs
			totalBytesTmp += rs.TotalBytes
		}

		break
	}

	if err != nil {
		return nil, metadata, totalBytes, err
	}

	logger.Infof("Restore successful, name: %s, result: %s", s.RepoName, utils.ToJSON(restoreSummarys))

	metadata = backupMetadata
	totalBytes = totalBytesTmp

	return restoreSummarys, metadata, totalBytes, nil
}
