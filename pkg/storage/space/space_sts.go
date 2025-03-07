package space

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"bytetrade.io/web3os/backups-sdk/pkg/constants"
	"bytetrade.io/web3os/backups-sdk/pkg/logger"
	"bytetrade.io/web3os/backups-sdk/pkg/utils"
	"github.com/emicklei/go-restful/v3"
	"github.com/pkg/errors"
)

type CloudStorageAccountResponse struct {
	Header
	Data *OlaresSpaceSession `json:"data"`
}

type Header struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type OlaresSpaceSession struct {
	Cloud          string `json:"cloud"`
	Bucket         string `json:"bucket"`
	Token          string `json:"st"`
	Prefix         string `json:"prefix"`
	Secret         string `json:"sk"`
	Key            string `json:"ak"`
	Expiration     string `json:"expiration"`
	Region         string `json:"region"`
	ResticRepo     string `json:"restic_repo"`
	ResticPassword string `json:"-"`
}

type StsToken struct {
	RepoName     string `json:"repo_name"`
	Storage      string `json:"storage"`
	Cloud        string `json:"cloud"`
	Region       string `json:"region"`
	Bucket       string `json:"bucket"`
	Prefix       string `json:"prefix"`
	AccessKey    string `json:"ak"`
	SecretKey    string `json:"sk"`
	SessionToken string `json:"st"`
	Expiration   string `json:"expiration"`
	ClusterId    string `json:"cluster_id"`
}

func (s *StsToken) RefreshStsToken(cloudApiMirror string) error {
	logger.Infof("refresh sts token")

	var url = s.getRequestSpaceRefreshStsUrl(cloudApiMirror)
	var headers = s.getRequestSpaceStsHeaders()
	var data = s.getRequestSpaceRefreshStsData()

	result, err := utils.Post[CloudStorageAccountResponse](url, headers, data)
	if err != nil {
		return err
	}

	queryResp := result

	if queryResp.Data == nil {
		return errors.WithStack(fmt.Errorf("get sts token invalid, code: %d, msg: %s, params: %s", queryResp.Code, queryResp.Message, data))
	}

	s.Cloud = queryResp.Data.Cloud
	s.Region = queryResp.Data.Region
	s.Bucket = queryResp.Data.Bucket
	s.Prefix = queryResp.Data.Prefix
	s.AccessKey = queryResp.Data.Key
	s.SecretKey = queryResp.Data.Secret
	s.SessionToken = queryResp.Data.Token
	s.Expiration = queryResp.Data.Expiration

	return nil
}

func (s *StsToken) GetStsToken(olaresDid, accessToken,
	cloudName, regionId, clusterId,
	cloudApiMirror string) error {
	logger.Info("get sts token")

	// ! test

	var url = s.getRequestSpaceStsUrl(cloudApiMirror)
	var headers = s.getRequestSpaceStsHeaders()
	var data = s.getRequestSpaceStsData(olaresDid, accessToken, cloudName, regionId, clusterId)

	result, err := utils.Post[CloudStorageAccountResponse](url, headers, data)
	if err != nil {
		return err
	}

	queryResp := result

	if queryResp.Code == 506 {
		return fmt.Errorf("user access token expired")
	}

	if queryResp.Data == nil {
		return errors.WithStack(fmt.Errorf("get sts token invalid, code: %d, msg: %s, data: %s", queryResp.Code, queryResp.Message, data))
	}

	s.Cloud = queryResp.Data.Cloud
	s.Region = queryResp.Data.Region
	s.Bucket = queryResp.Data.Bucket
	s.Prefix = queryResp.Data.Prefix
	s.AccessKey = queryResp.Data.Key
	s.SecretKey = queryResp.Data.Secret
	s.SessionToken = queryResp.Data.Token
	s.Expiration = queryResp.Data.Expiration

	return nil
}

func (s *StsToken) parseClusterId(clusterId string) string {
	return utils.Base64encode([]byte(clusterId))
}

func (s *StsToken) parseSpaceStsDuration() time.Duration {
	debugCustomStsDuration := os.Getenv("DEBUG_STS_TOKEN_DURATION")
	if debugCustomStsDuration != "" {
		var dur = 15 * time.Minute
		return dur
	}
	return 12 * time.Hour
}

func (s *StsToken) getRequestSpaceStsUrl(cloudApiMirror string) string {
	return fmt.Sprintf("%s%s", cloudApiMirror, constants.StsTokenUrl)
}

func (s *StsToken) getRequestSpaceStsHeaders() map[string]string {
	var headers = make(map[string]string)
	headers[restful.HEADER_ContentType] = "application/x-www-form-urlencoded"

	return headers
}

func (s *StsToken) getRequestSpaceStsData(olaresDid, token, location, region, clusterId string) string {
	var data = fmt.Sprintf("cloudName=%s&durationSeconds=%s&region=%s&token=%s&userid=%s&clusterId=%s",
		location, fmt.Sprintf("%.0f", s.parseSpaceStsDuration().Seconds()), region, token, olaresDid, s.parseClusterId(clusterId))
	return data
}

func (s *StsToken) getRequestSpaceRefreshStsUrl(cloudApiMirror string) string {
	return fmt.Sprintf("%s%s", cloudApiMirror, constants.StsTokenRefreshUrl)
}

func (s *StsToken) getRequestSpaceRefreshStsData() string {
	var sk = url.QueryEscape(s.SecretKey)
	var st = url.QueryEscape(s.SessionToken)
	var data = fmt.Sprintf("ak=%s&sk=%s&st=%s&durationSeconds=%s",
		s.AccessKey, sk, st, fmt.Sprintf("%.0f", s.parseSpaceStsDuration().Seconds()))

	return data
}
