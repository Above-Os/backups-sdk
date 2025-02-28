package space

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"bytetrade.io/web3os/backups-sdk/pkg/client"
	"bytetrade.io/web3os/backups-sdk/pkg/common"
	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/util"
	"bytetrade.io/web3os/backups-sdk/pkg/util/logger"
	"bytetrade.io/web3os/backups-sdk/pkg/util/net"
	"github.com/emicklei/go-restful/v3"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/util/retry"
)

var debugDuration = true

type OlaresSpaceParam struct {
	Duration  string
	Region    string
	ClusterId string
}

type SpaceToken struct {
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

func (s *SpaceToken) IsSpaceTokenValid(repoName, repoRegion string) bool {
	return s.isSpaceKeyValid(repoName, repoRegion) && !s.isSpaceTokenExpired()
}

func (s *SpaceToken) RefreshSpaceToken(spaceUserName, cloudApiMirror string) error {
	var backoff = wait.Backoff{
		Duration: 3 * time.Second,
		Factor:   2,
		Jitter:   0.1,
		Steps:    2,
	}

	logger.Infof("refresh space %s token, cluster: %s", spaceUserName)
	var err = retry.OnError(backoff, func(err error) bool {
		return true
	}, func() error {
		var url = s.getRequestSpaceRefreshTokenUrl(cloudApiMirror)
		var headers = s.getRequestSpaceTokenHeaders()
		var data = s.getRequestSpaceRefreshTokenData()

		result, err := net.Post[CloudStorageAccountResponse](url, headers, data, true, true)
		if err != nil {
			return err
		}

		queryResp := result
		if queryResp.Code != http.StatusOK { // 501(missing params) 506(expired or un-login)
			return errors.WithStack(fmt.Errorf("refresh space token account from cloud error: %d, data: %s",
				queryResp.Code, queryResp.Message))
		}

		if queryResp.Data == nil {
			return errors.WithStack(fmt.Errorf("refresh space token account from cloud data is empty, code: %d, data: %s", queryResp.Code, queryResp.Message))
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
	})

	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (s *SpaceToken) GetSpaceToken(olaresDid, olaresId, olaresName, spaceUserAccessToken,
	spaceLocation, spaceRegion,
	cloudApiMirror string) error {
	var clusterId, err = s.getClusterId()
	if err != nil {
		return err
	}

	var backoff = wait.Backoff{
		Duration: 3 * time.Second,
		Factor:   2,
		Jitter:   0.1,
		Steps:    3,
	}

	logger.Infof("get space %s token, cluster: %s", olaresName, clusterId)
	err = retry.OnError(backoff, func(err error) bool {
		return true
	}, func() error {
		var url = s.getRequestSpaceTokenUrl(cloudApiMirror)
		var headers = s.getRequestSpaceTokenHeaders()
		var data = s.getRequestSpaceTokenData(olaresDid, spaceUserAccessToken, spaceLocation, spaceRegion, clusterId)

		result, err := net.Post[CloudStorageAccountResponse](url, headers, data, true, true)
		if err != nil {
			return err
		}

		queryResp := result

		if queryResp.Data == nil {
			return errors.WithStack(fmt.Errorf("get cloud storage account from cloud data is empty, code: %d, data: %s", queryResp.Code, queryResp.Message))
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
	})

	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (s *SpaceToken) GetEnv(repoName string, password string) *restic.ResticEnv {
	repo, _ := s.formatSpaceRepository(repoName)

	var envs = &restic.ResticEnv{
		AWS_ACCESS_KEY_ID:     s.AccessKey,
		AWS_SECRET_ACCESS_KEY: s.SecretKey,
		AWS_SESSION_TOKEN:     s.SessionToken,
		RESTIC_REPOSITORY:     repo,
		RESTIC_PASSWORD:       password,
	}

	return envs
}

//
//
//

func (s *SpaceToken) formatSpaceRepository(repoName string) (repository string, err error) {
	var repoPrefix = filepath.Join(s.Prefix, "restic", repoName)
	var domain = fmt.Sprintf("s3.%s.%s", s.Region, common.AwsDomain)
	var repo = filepath.Join(domain, s.Bucket, repoPrefix)
	repository = fmt.Sprintf("s3:%s", repo)
	return
}

func (s *SpaceToken) getClusterId() (clusterId string, err error) {
	var factory client.Factory
	factory, err = client.NewFactory()
	if err != nil {
		return
	}

	var dynamicClient dynamic.Interface
	dynamicClient, err = factory.DynamicClient()
	if err != nil {
		return
	}

	var backoff = wait.Backoff{
		Duration: 2 * time.Second,
		Factor:   2,
		Jitter:   0.1,
		Steps:    5,
	}

	var resourceName = "terminus"
	var gvr = schema.GroupVersionResource{
		Group:    "sys.bytetrade.io",
		Version:  "v1alpha1",
		Resource: resourceName,
	}

	if err := retry.OnError(backoff, func(err error) bool {
		return true
	}, func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		unstructuredUser, err := dynamicClient.Resource(gvr).Get(ctx, resourceName, metav1.GetOptions{})
		if err != nil {
			return errors.WithStack(err)
		}
		obj := unstructuredUser.UnstructuredContent()
		clusterId, _, err = unstructured.NestedString(obj, "metadata", "labels", "bytetrade.io/cluster-id")
		if err != nil {
			return errors.WithStack(err)
		}
		if clusterId == "" {
			return errors.WithStack(fmt.Errorf("cluster id not found"))
		}
		return nil
	}); err != nil {
		err = errors.WithStack(fmt.Errorf("get cluster id error %v", err))
	}

	return
}

func (s *SpaceToken) parseSpaceRegion() string {
	return "us-east-1"
}

func (s *SpaceToken) parseSpaceLocation() string {
	return "aws"
}

func (s *SpaceToken) parseClusterId(clusterId string) string {
	return util.Base64encode([]byte(clusterId))
}

func (s *SpaceToken) parseSpaceTokenDuration(isDebug bool) time.Duration {
	if isDebug {
		var dur = 15 * time.Minute
		return dur
	}
	return 12 * time.Hour
}

func (s *SpaceToken) isSpaceKeyValid(repoName, repoRegion string) bool {
	if s.RepoName == "" || s.Region == "" || s.AccessKey == "" || s.SecretKey == "" || s.SessionToken == "" || s.RepoName != repoName || s.Region != repoRegion {
		return false
	}
	return true
}

func (s *SpaceToken) isSpaceTokenExpired() bool {
	expiration, err := strconv.ParseInt(s.Expiration, 10, 64)
	if err != nil {
		return true
	}
	expiresTime, expired := util.IsTimestampAboutToExpire(expiration)
	logger.Infof("space access token expires at %s", expiresTime.String())
	return expired
}

func (s *SpaceToken) getRequestSpaceTokenUrl(cloudApiMirror string) string {
	return fmt.Sprintf("%s/v1/resource/stsToken/backup", s.getCloudApi(cloudApiMirror))
}

func (s *SpaceToken) getRequestSpaceTokenHeaders() map[string]string {
	var headers = make(map[string]string)
	headers[restful.HEADER_ContentType] = "application/x-www-form-urlencoded"

	return headers
}

func (s *SpaceToken) getRequestSpaceTokenData(olaresDid, token, location, region, clusterId string) string {
	var data = fmt.Sprintf("cloudName=%s&durationSeconds=%s&region=%s&token=%s&userid=%s&clusterId=%s",
		location, fmt.Sprintf("%.0f", s.parseSpaceTokenDuration(debugDuration).Seconds()), region, token, olaresDid, s.parseClusterId(clusterId))
	return data
}

func (s *SpaceToken) getRequestSpaceRefreshTokenUrl(cloudApiMirror string) string {
	return fmt.Sprintf("%s/v1/resource/stsToken/backup/refresh", s.getCloudApi(cloudApiMirror))
}

func (s *SpaceToken) getRequestSpaceRefreshTokenData() string {
	var data = fmt.Sprintf("ak=%s&sk=%s&st=%s&durationSeconds=%s",
		s.AccessKey, s.SecretKey, s.SessionToken, fmt.Sprintf("%.0f", s.parseSpaceTokenDuration(debugDuration).Seconds()))

	return data
}

func (s *SpaceToken) getCloudApi(cloudApiMirror string) string {
	var serverDomain = util.DefaultValue(common.DefaultCloudApiUrl, cloudApiMirror)
	return strings.TrimRight(serverDomain, "/")
}
