package space

import (
	"fmt"
	"path/filepath"
	"strings"

	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/util"
	"bytetrade.io/web3os/backups-sdk/pkg/util/net"
	"github.com/emicklei/go-restful/v3"
	"github.com/pkg/errors"
)

const (
	SpaceDomain = "amazonaws.com"
)

type Space struct {
	RepoName        string
	SnapshotId      string
	OlaresDid       string
	AccessToken     string
	ClusterId       string
	CloudName       string
	RegionId        string
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

type Regions []*Region

type CloudStorageRegionResponse struct {
	Header
	Data Regions `json:"data"`
}

type Region struct {
	RegionId   string `json:"regionId"`
	RegionName string `json:"regionName"`
	CloudName  string `json:"cloudName"`
}

func (s *Space) Regions() error {
	var url = fmt.Sprintf("%s/v1/resource/backup/region", s.getCloudApi())
	var headers = map[string]string{
		restful.HEADER_ContentType: "application/x-www-form-urlencoded",
	}
	var data = fmt.Sprintf("userid=%s&token=%s", s.OlaresDid, s.AccessToken)

	result, err := net.Post[CloudStorageRegionResponse](url, headers, data)
	if err != nil {
		return err
	}

	if result.Data == nil {
		return errors.WithStack(fmt.Errorf("get regions invalid, code: %d, msg: %s, params: %s", result.Code, result.Message, data))
	}

	return nil
}

func (s *Space) GetEnv(repoName string) *restic.ResticEnvs {
	s.RepoName = repoName
	repo, _ := s.FormatRepository()

	var envs = &restic.ResticEnvs{
		AWS_ACCESS_KEY_ID:     s.StsToken.AccessKey,
		AWS_SECRET_ACCESS_KEY: s.StsToken.SecretKey,
		AWS_SESSION_TOKEN:     s.StsToken.SessionToken,
		RESTIC_REPOSITORY:     repo,
		RESTIC_PASSWORD:       s.Password,
	}

	return envs
}

func (s *Space) getCosRepository() (repository string, err error) {
	var repoPrefix = filepath.Join(s.StsToken.Prefix, "restic", s.RepoName)
	repository = fmt.Sprintf("s3:https://cos.%s.myqcloud.com/%s/%s%s", s.RegionId, s.StsToken.Bucket, repoPrefix, s.RepoName)
	return
}

func (s *Space) getDefaultRepository() (repository string, err error) {
	var repoPrefix = filepath.Join(s.StsToken.Prefix, "restic", s.RepoName)
	var domain = fmt.Sprintf("s3.%s.%s", s.StsToken.Region, SpaceDomain)
	var repo = filepath.Join(domain, s.StsToken.Bucket, repoPrefix)
	repository = fmt.Sprintf("s3:%s", repo)
	return
}

func (s *Space) FormatRepository() (repository string, err error) {
	if s.CloudName == "TencentCloud" {
		return s.getCosRepository()
	} else {
		return s.getDefaultRepository()
	}
}

func (s *Space) getStsToken(repoLocation, repoRegion string) error {
	if err := s.StsToken.GetStsToken(s.OlaresDid, s.AccessToken, repoLocation, repoRegion, s.ClusterId, s.getCloudApi()); err != nil {
		return errors.WithStack(fmt.Errorf("get sts token error: %v", err))
	}
	return nil
}

func (s *Space) refreshStsTokens() error {
	if err := s.StsToken.RefreshStsToken(s.getCloudApi()); err != nil {
		return errors.WithStack(fmt.Errorf("refresh sts token error: %v", err))
	}

	return nil
}

func (s *Space) getCloudApi() string {
	var serverDomain = util.DefaultValue(DefaultCloudApiUrl, s.CloudApiMirror)
	return strings.TrimRight(serverDomain, "/")
}
