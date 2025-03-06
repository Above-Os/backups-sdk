package options

import (
	"github.com/spf13/cobra"
)

// ~ space
var _ Option = &SpaceSnapshotsOption{}

type SpaceSnapshotsOption struct {
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

// ~ s3
var _ Option = &S3SnapshotsOption{}

type S3SnapshotsOption struct {
	RepoName        string
	Endpoint        string
	AccessKey       string
	SecretAccessKey string
}

func NewSnapshotsS3Option() *S3SnapshotsOption {
	return &S3SnapshotsOption{}
}

func (o *S3SnapshotsOption) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.RepoName, "repo-name", "", "", "Backup repo name")
	cmd.Flags().StringVarP(&o.Endpoint, "endpoint", "", "", "Endpoint for S3, for example https://{bucket}.{region}.amazonaws.com/{prefix}")
	cmd.Flags().StringVarP(&o.AccessKey, "access-key", "", "", "Access Key for S3")
	cmd.Flags().StringVarP(&o.SecretAccessKey, "secret-access-key", "", "", "Secret Access Key for S3")
}

// ~ cos
var _ Option = &CosSnapshotsOption{}

type CosSnapshotsOption struct {
	RepoName        string
	Endpoint        string
	AccessKey       string
	SecretAccessKey string
}

func NewSnapshotsCosOption() *CosSnapshotsOption {
	return &CosSnapshotsOption{}
}

func (o *CosSnapshotsOption) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.RepoName, "repo-name", "", "", "Backup repo name")
	cmd.Flags().StringVarP(&o.Endpoint, "endpoint", "", "", "Endpoint for Tencent COS, for example https://cos.{region}.myqcloud.com/{bucket}/{prefix}")
	cmd.Flags().StringVarP(&o.AccessKey, "access-key", "", "", "Access Key for Tencent COS")
	cmd.Flags().StringVarP(&o.SecretAccessKey, "secret-access-key", "", "", "Secret Access Key for Tencent COS")
}

// ~ filesystem
var _ Option = &FilesystemSnapshotsOption{}

type FilesystemSnapshotsOption struct {
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
