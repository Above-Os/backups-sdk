package options

import (
	"github.com/spf13/cobra"
)

// ~ space
var _ Option = &SpaceRestoreOption{}

type SpaceRestoreOption struct {
	RepoName       string
	SnapshotId     string
	Path           string
	OlaresDid      string
	AccessToken    string
	ClusterId      string
	CloudName      string
	RegionId       string
	CloudApiMirror string
}

func NewRestoreSpaceOption() *SpaceRestoreOption {
	return &SpaceRestoreOption{}
}

func (o *SpaceRestoreOption) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.RepoName, "repo-name", "", "", "Backup repo name")
	cmd.Flags().StringVarP(&o.SnapshotId, "snapshot-id", "", "", "Snapshot ID")
	cmd.Flags().StringVarP(&o.Path, "path", "", "", "The directory to be restore")
	cmd.Flags().StringVarP(&o.OlaresDid, "olares-did", "", "", "Olares DID")
	cmd.Flags().StringVarP(&o.AccessToken, "access-token", "", "", "Space Access Token")
	cmd.Flags().StringVarP(&o.ClusterId, "cluster-id", "", "", "Olares Cluster ID")
	cmd.Flags().StringVarP(&o.CloudName, "region", "", "", "Space Cloud Name")
	cmd.Flags().StringVarP(&o.RegionId, "region", "", "", "Space Region Id")
	cmd.Flags().StringVarP(&o.CloudApiMirror, "cloud-api-mirror", "", "", "Cloud API mirror")
}

// ~ s3
var _ Option = &S3RestoreOption{}

type S3RestoreOption struct {
	RepoName        string
	SnapshotId      string
	Endpoint        string
	AccessKey       string
	SecretAccessKey string
	Path            string
}

func NewRestoreS3Option() *S3RestoreOption {
	return &S3RestoreOption{}
}

func (o *S3RestoreOption) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.RepoName, "repo-name", "", "", "Backup repo name")
	cmd.Flags().StringVarP(&o.SnapshotId, "snapshot-id", "", "", "Snapshot ID")
	cmd.Flags().StringVarP(&o.Endpoint, "endpoint", "", "", "Endpoint for S3, for example https://{bucket}.{region}.amazonaws.com/{prefix}")
	cmd.Flags().StringVarP(&o.AccessKey, "access-key", "", "", "Access Key for S3")
	cmd.Flags().StringVarP(&o.SecretAccessKey, "secret-access-key", "", "", "Secret Access Key for S3")
	cmd.Flags().StringVarP(&o.Path, "path", "", "", "The directory to be restore")
}

// ~ cos
var _ Option = &CosRestoreOption{}

type CosRestoreOption struct {
	RepoName        string
	SnapshotId      string
	Endpoint        string
	AccessKey       string
	SecretAccessKey string
	Path            string
}

func NewRestoreCosOption() *CosRestoreOption {
	return &CosRestoreOption{}
}

func (o *CosRestoreOption) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.RepoName, "repo-name", "", "", "Backup repo name")
	cmd.Flags().StringVarP(&o.SnapshotId, "snapshot-id", "", "", "Snapshot ID")
	cmd.Flags().StringVarP(&o.Endpoint, "endpoint", "", "", "Endpoint for Tencent COS, for example https://cos.{region}.myqcloud.com/{bucket}/{prefix}")
	cmd.Flags().StringVarP(&o.AccessKey, "access-key", "", "", "Access Key for Tencent COS")
	cmd.Flags().StringVarP(&o.SecretAccessKey, "secret-access-key", "", "", "Secret Access Key for Tencent COS")
	cmd.Flags().StringVarP(&o.Path, "path", "", "", "The directory to be restore")
}

// ~ filesystem
var _ Option = &FilesystemRestoreOption{}

type FilesystemRestoreOption struct {
	RepoName   string
	SnapshotId string
	Endpoint   string
	Path       string
	OlaresId   string
}

func NewRestoreFilesystemOption() *FilesystemRestoreOption {
	return &FilesystemRestoreOption{}
}

func (o *FilesystemRestoreOption) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.RepoName, "repo-name", "", "", "Backup repo name")
	cmd.Flags().StringVarP(&o.SnapshotId, "snapshot-id", "", "", "Snapshot ID")
	cmd.Flags().StringVarP(&o.Endpoint, "endpoint", "", "", "The endpoint of the filesystem is the local computer directory where the backup will be stored")
	cmd.Flags().StringVarP(&o.Path, "path", "", "", "The directory to be restore")
	cmd.Flags().StringVarP(&o.OlaresId, "olares-id", "", "", "Olares ID")
}
