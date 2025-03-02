package cos

import (
	"context"

	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/util"
	"bytetrade.io/web3os/backups-sdk/pkg/util/logger"
)

func (c *Cos) Backup() error {
	repository, err := c.FormatRepository()
	if err != nil {
		return err
	}

	var resticEnv = c.GetEnv(repository)

	logger.Debugf("cos backup env vars: %s", util.Base64encode([]byte(resticEnv.ToString())))

	r, err := restic.NewRestic(context.Background(), c.RepoName, "", resticEnv.ToMap(), nil)
	if err != nil {
		return err
	}

	_, initRepo, err := r.Init()
	if err != nil {
		return err
	}

	if !initRepo {
		if err = r.Repair(); err != nil {
			return err
		}
	}

	backupResult, err := r.Backup(c.RepoName, c.Path, "")
	if err != nil {
		return err
	}
	logger.Infof("Backup to Tencent COS success, result id: %s", util.ToJSON(backupResult))

	return nil
}
