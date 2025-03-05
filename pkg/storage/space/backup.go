package space

import (
	"context"
	"fmt"

	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/util"
	"bytetrade.io/web3os/backups-sdk/pkg/util/logger"
	"github.com/pkg/errors"
)

func (s *Space) Backup() (err error) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	if err = s.getStsToken(s.CloudName, s.RegionId); err != nil {
		return errors.WithStack(err)
	}

	var backupSummary *restic.SummaryOutput

	for {
		var initResult string
		var initialized bool

		var envs = s.GetEnv(s.RepoName)
		var opts = &restic.ResticOptions{
			RepoName:        s.RepoName,
			RepoEnvs:        envs,
			LimitUploadRate: s.LimitUploadRate,
		}

		logger.Debugf("space backup env vars: %s", util.Base64encode([]byte(envs.String())))

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
			logger.Infof("repo %s already initialized", s.RepoName)
			logger.Infof("repairing repo %s index", s.RepoName)
			if err = r.Repair(); err != nil {
				break
			}
		} else {
			logger.Infof("repo %s initialized\n\n%s", s.RepoName, initResult)
		}

		logger.Infof("preparing to start repo %s backup", s.RepoName)

		backupSummary, err = r.Backup(s.Path, "")
		if err != nil {
			switch err.Error() {
			case restic.ERROR_MESSAGE_TOKEN_EXPIRED.Error():
				logger.Infof("space backup upload stopped, sts token expired, refresh and retring...")
				if err = s.refreshStsTokens(); err != nil {
					err = fmt.Errorf("space backup upload sts token service refresh-token error: %v", err)
					break
				}
				continue
			default:
				break
			}
		}
		break
	}

	if err != nil {
		return errors.WithStack(err)
	}

	fmt.Println("Backup successful, result: ", util.ToJSON(backupSummary))

	return nil
}
