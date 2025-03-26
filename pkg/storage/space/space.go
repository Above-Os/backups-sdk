package space

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"bytetrade.io/web3os/backups-sdk/pkg/constants"
	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/model"
	"bytetrade.io/web3os/backups-sdk/pkg/utils"
	"github.com/emicklei/go-restful/v3"
	"github.com/pkg/errors"
)

type Space struct {
	RepoName          string
	SnapshotId        string
	OlaresDid         string
	AccessToken       string
	ClusterId         string
	CloudName         string
	RegionId          string
	Password          string
	Path              string
	LimitUploadRate   string
	LimitDownloadRate string
	CloudApiMirror    string
	StsToken          *StsToken
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

func (s *Space) Regions() ([]map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var url = fmt.Sprintf("%s/v1/resource/backup/region", s.getCloudApi())
	var headers = map[string]string{
		restful.HEADER_ContentType: "application/x-www-form-urlencoded",
	}
	var data = fmt.Sprintf("userid=%s&token=%s", s.OlaresDid, s.AccessToken)

	result, err := utils.Post[CloudStorageRegionResponse](ctx, url, headers, data)
	if err != nil {
		return nil, err
	}

	if result.Data == nil {
		return nil, errors.WithStack(fmt.Errorf("get regions invalid, code: %d, msg: %s, params: %s", result.Code, result.Message, data))
	}

	var regions []map[string]string

	for _, region := range result.Data {
		var r = make(map[string]string)
		r["cloudName"] = region.CloudName
		r["regionId"] = region.RegionId
		regions = append(regions, r)
	}

	return regions, nil
}

func (s *Space) GetEnv(repository string) *restic.ResticEnvs {
	var envs = &restic.ResticEnvs{
		AWS_ACCESS_KEY_ID:     s.StsToken.AccessKey,
		AWS_SECRET_ACCESS_KEY: s.StsToken.SecretKey,
		AWS_SESSION_TOKEN:     s.StsToken.SessionToken,
		RESTIC_REPOSITORY:     repository,
		RESTIC_PASSWORD:       s.Password,
	}

	return envs
}

func (s *Space) getCosRepository() (storageInfo *model.StorageInfo, err error) {
	var repoPrefix = filepath.Join(s.StsToken.Prefix, "restic", s.RepoName)
	var repository = fmt.Sprintf("s3:https://cos.%s.%s/%s/%s", s.RegionId, constants.StorageTencentDoman, s.StsToken.Bucket, repoPrefix)

	storageInfo = &model.StorageInfo{
		Location:  "space",
		Url:       repository,
		CloudName: s.CloudName,
		RegionId:  s.RegionId,
		Bucket:    s.StsToken.Bucket,
		Prefix:    s.StsToken.Prefix,
	}

	return
}

func (s *Space) getDefaultRepository() (storageInfo *model.StorageInfo, err error) {
	var repoPrefix = filepath.Join(s.StsToken.Prefix, "restic", s.RepoName)
	var domain = fmt.Sprintf("s3.%s.%s", s.StsToken.Region, constants.StorageS3Domain)
	var repo = filepath.Join(domain, s.StsToken.Bucket, repoPrefix)
	var repository = fmt.Sprintf("s3:%s", repo)

	storageInfo = &model.StorageInfo{
		Location:  "space",
		Url:       repository,
		CloudName: s.CloudName,
		RegionId:  s.RegionId,
		Bucket:    s.StsToken.Bucket,
		Prefix:    s.StsToken.Prefix,
	}

	return
}

func (s *Space) FormatRepository() (storageInfo *model.StorageInfo, err error) {
	if s.CloudName == constants.CloudTencentName {
		return s.getCosRepository()
	} else {
		return s.getDefaultRepository()
	}
}

func (s *Space) getStsToken(ctx context.Context) error {
	if err := s.StsToken.GetStsToken(ctx, s.OlaresDid, s.AccessToken, s.CloudName, s.RegionId, s.ClusterId, s.getCloudApi()); err != nil {
		return errors.WithStack(fmt.Errorf("get sts token error: %v", err))
	}
	return nil
}

func (s *Space) refreshStsTokens(ctx context.Context) error {
	if err := s.StsToken.RefreshStsToken(ctx, s.getCloudApi()); err != nil {
		return errors.WithStack(fmt.Errorf("refresh sts token error: %v", err))
	}

	return nil
}

func (s *Space) getCloudApi() string {
	var serverDomain = utils.DefaultValue(constants.DefaultCloudApiUrl, s.CloudApiMirror)
	return strings.TrimRight(serverDomain, "/")
}
