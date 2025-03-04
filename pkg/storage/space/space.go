package space

import (
	"fmt"
	"path/filepath"

	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"github.com/pkg/errors"
)

const (
	SpaceDomain     = "amazonaws.com"
	DefaultLocation = "aws"
	DefaultRegion   = "us-east-1"
)

type Space struct {
	RepoName        string
	SnapshotId      string
	RepoRegion      string
	OlaresDid       string
	AccessToken     string
	ClusterId       string
	Password        string
	Path            string
	LimitUploadRate string
	CloudApiMirror  string
	StsToken        *StsToken
}

type StorageResponse struct {
	BackupSummary    *restic.SummaryOutput
	RestoreSummary   *restic.RestoreSummaryOutput
	SnapshotsSummary []*restic.Snapshot
	Error            error
}

func (s *Space) GetEnv(repoName string) *restic.ResticEnv {
	s.RepoName = repoName
	repo, _ := s.FormatRepository()

	var envs = &restic.ResticEnv{
		AWS_ACCESS_KEY_ID:     s.StsToken.AccessKey,
		AWS_SECRET_ACCESS_KEY: s.StsToken.SecretKey,
		AWS_SESSION_TOKEN:     s.StsToken.SessionToken,
		RESTIC_REPOSITORY:     repo,
		RESTIC_PASSWORD:       s.Password,
	}

	return envs
}
func (s *Space) FormatRepository() (repository string, err error) {
	var repoPrefix = filepath.Join(s.StsToken.Prefix, "restic", s.RepoName)
	var domain = fmt.Sprintf("s3.%s.%s", s.StsToken.Region, SpaceDomain)
	var repo = filepath.Join(domain, s.StsToken.Bucket, repoPrefix)
	repository = fmt.Sprintf("s3:%s", repo)
	return
}

func (s *Space) getStsToken(repoLocation, repoRegion string) error {
	if err := s.StsToken.GetStsToken(s.OlaresDid, s.AccessToken, repoLocation, repoRegion, s.ClusterId, s.CloudApiMirror); err != nil {
		return errors.WithStack(fmt.Errorf("get sts token error: %v", err))
	}
	return nil
}

func (s *Space) refreshStsTokens() error {
	if err := s.StsToken.RefreshStsToken(s.CloudApiMirror); err != nil {
		return errors.WithStack(fmt.Errorf("refresh sts token error: %v", err))
	}

	return nil
}
