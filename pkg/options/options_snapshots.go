package options

import (
	"github.com/spf13/cobra"
)

// ~ space
var _ Option = &SpaceSnapshotsOption{}

type SpaceSnapshotsOption struct {
	RepoId         string
	RepoName       string
	OlaresDid      string
	AccessToken    string
	ClusterId      string
	CloudName      string
	RegionId       string
	CloudApiMirror string
}

func NewSnapshotsSpaceOption() *SpaceSnapshotsOption {
	return &SpaceSnapshotsOption{}
}

func (o *SpaceSnapshotsOption) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.RepoName, "repo-name", "", "", "Backup repo name")
	cmd.Flags().StringVarP(&o.OlaresDid, "olares-did", "", "", "Olares DID")
	cmd.Flags().StringVarP(&o.AccessToken, "access-token", "", "", "Space Access Token")
	cmd.Flags().StringVarP(&o.ClusterId, "cluster-id", "", "", "Space Cluster ID")
	cmd.Flags().StringVarP(&o.CloudName, "cloud-name", "", "", "Space Cloud Name")
	cmd.Flags().StringVarP(&o.RegionId, "region-id", "", "", "Space Region Id")
	cmd.Flags().StringVarP(&o.CloudApiMirror, "cloud-api-mirror", "", "", "Cloud API mirror")
}

// ~ aws
var _ Option = &AwsSnapshotsOption{}

type AwsSnapshotsOption struct {
	RepoId          string
	RepoName        string
	Endpoint        string
	AccessKey       string
	SecretAccessKey string
}

func NewSnapshotsAwsOption() *AwsSnapshotsOption {
	return &AwsSnapshotsOption{}
}

func (o *AwsSnapshotsOption) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.RepoName, "repo-name", "", "", "Backup repo name")
	cmd.Flags().StringVarP(&o.Endpoint, "endpoint", "", "", "Endpoint for S3, for example https://{bucket}.{region}.amazonaws.com/{prefix}")
	cmd.Flags().StringVarP(&o.AccessKey, "access-key", "", "", "Access Key for S3")
	cmd.Flags().StringVarP(&o.SecretAccessKey, "secret-access-key", "", "", "Secret Access Key for S3")
}

// ~ cos
var _ Option = &TencentCloudSnapshotsOption{}

type TencentCloudSnapshotsOption struct {
	RepoId          string
	RepoName        string
	Endpoint        string
	AccessKey       string
	SecretAccessKey string
}

func NewSnapshotsTencentCloudOption() *TencentCloudSnapshotsOption {
	return &TencentCloudSnapshotsOption{}
}

func (o *TencentCloudSnapshotsOption) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.RepoName, "repo-name", "", "", "Backup repo name")
	cmd.Flags().StringVarP(&o.Endpoint, "endpoint", "", "", "Endpoint for Tencent COS, for example https://cos.{region}.myqcloud.com/{bucket}/{prefix}")
	cmd.Flags().StringVarP(&o.AccessKey, "access-key", "", "", "Access Key for Tencent COS")
	cmd.Flags().StringVarP(&o.SecretAccessKey, "secret-access-key", "", "", "Secret Access Key for Tencent COS")
}

// ~ filesystem
var _ Option = &FilesystemSnapshotsOption{}

type FilesystemSnapshotsOption struct {
	RepoId   string
	RepoName string
	Endpoint string
}

func NewSnapshotsFilesystemOption() *FilesystemSnapshotsOption {
	return &FilesystemSnapshotsOption{}
}

func (o *FilesystemSnapshotsOption) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.RepoName, "repo-name", "", "", "Backup repo name")
	cmd.Flags().StringVarP(&o.Endpoint, "endpoint", "", "", "The endpoint of the filesystem is the local computer directory where the backup will be stored")
}
