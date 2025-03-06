package options

import (
	"github.com/spf13/cobra"
)

// ~ space
var _ Option = &SpaceBackupOption{}

type SpaceBackupOption struct {
	RepoName        string `json:"repo_name"`
	Path            string `json:"path"`
	LimitUploadRate string `json:"limit_upload_rate"`
	OlaresDid       string `json:"olares_did"`
	AccessToken     string `json:"access_token"`
	ClusterId       string `json:"cluster_id"`
	CloudName       string `json:"cloud_name"`
	RegionId        string `json:"region_id"`
	CloudApiMirror  string `json:"cloud_api_mirror"`
}

func NewBackupSpaceOption() *SpaceBackupOption {
	return &SpaceBackupOption{}
}

func (o *SpaceBackupOption) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.RepoName, "repo-name", "", "", "Backup repo name")
	cmd.Flags().StringVarP(&o.Path, "path", "", "", "The directory to be backed up")
	cmd.Flags().StringVarP(&o.LimitUploadRate, "limit-upload-rate", "", "", "Limits uploads to a maximum rate in KiB/s. (default: unlimited)")
	cmd.Flags().StringVarP(&o.OlaresDid, "olares-did", "", "", "Olares DID")
	cmd.Flags().StringVarP(&o.AccessToken, "access-token", "", "", "Space Access Token")
	cmd.Flags().StringVarP(&o.ClusterId, "cluster-id", "", "", "Space Cluster ID")
	cmd.Flags().StringVarP(&o.CloudName, "cloud-name", "", "", "Space Cloud Name")
	cmd.Flags().StringVarP(&o.RegionId, "region-id", "", "", "Space Region Id")
	cmd.Flags().StringVarP(&o.CloudApiMirror, "cloud-api-mirror", "", "", "Cloud API mirror")
}

// ~ s3
var _ Option = &S3BackupOption{}

type S3BackupOption struct {
	RepoName        string
	Endpoint        string
	AccessKey       string
	SecretAccessKey string
	Path            string
	LimitUploadRate string
}

func NewBackupS3Option() *S3BackupOption {
	return &S3BackupOption{}
}

func (o *S3BackupOption) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.RepoName, "repo-name", "", "", "Backup repo name")

	cmd.Flags().StringVarP(&o.Endpoint, "endpoint", "", "", "Endpoint for S3, for example https://{bucket}.{region}.amazonaws.com/{prefix}")
	cmd.Flags().StringVarP(&o.AccessKey, "access-key", "", "", "Access Key for S3")
	cmd.Flags().StringVarP(&o.SecretAccessKey, "secret-access-key", "", "", "Secret Access Key for S3")

	cmd.Flags().StringVarP(&o.Path, "path", "", "", "The directory to be backed up")
	cmd.Flags().StringVarP(&o.LimitUploadRate, "limit-upload-rate", "", "", "Limits uploads to a maximum rate in KiB/s. (default: unlimited)")
}

// ~ cos
var _ Option = &CosBackupOption{}

type CosBackupOption struct {
	RepoName        string
	Endpoint        string
	AccessKey       string
	SecretAccessKey string
	Path            string
	LimitUploadRate string
}

func NewBackupCosOption() *CosBackupOption {
	return &CosBackupOption{}
}

func (o *CosBackupOption) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.RepoName, "repo-name", "", "", "Backup repo name")

	cmd.Flags().StringVarP(&o.Endpoint, "endpoint", "", "", "Endpoint for Tencent COS, for example https://cos.{region}.myqcloud.com/{bucket}/{prefix}")
	cmd.Flags().StringVarP(&o.AccessKey, "access-key", "", "", "Access Key for Tencent COS")
	cmd.Flags().StringVarP(&o.SecretAccessKey, "secret-access-key", "", "", "Secret Access Key for Tencent COS")

	cmd.Flags().StringVarP(&o.Path, "path", "", "", "The directory to be backed up")
	cmd.Flags().StringVarP(&o.LimitUploadRate, "limit-upload-rate", "", "", "Limits uploads to a maximum rate in KiB/s. (default: unlimited)")
}

// ~ filesystem
var _ Option = &FilesystemBackupOption{}

type FilesystemBackupOption struct {
	RepoName string
	Endpoint string
	Path     string
}

func NewBackupFilesystemOption() *FilesystemBackupOption {
	return &FilesystemBackupOption{}
}

func (o *FilesystemBackupOption) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.RepoName, "repo-name", "", "", "Backup repo name")
	cmd.Flags().StringVarP(&o.Endpoint, "endpoint", "", "", "The endpoint of the filesystem is the local computer directory where the backup will be stored")
	cmd.Flags().StringVarP(&o.Path, "path", "", "", "The directory to be backed up")
}
