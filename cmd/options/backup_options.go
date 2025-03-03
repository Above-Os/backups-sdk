package options

import (
	"bytetrade.io/web3os/backups-sdk/pkg/common"
	"github.com/spf13/cobra"
)

// ~ space
type SpaceBackupOption struct {
	RepoName        string
	Path            string
	LimitUploadRate string
	OlaresId        string
	BaseDir         string
	CloudApiMirror  string
}

func NewBackupSpaceOption() *SpaceBackupOption {
	return &SpaceBackupOption{}
}

func (o *SpaceBackupOption) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.RepoName, "repo-name", "", "", "Backup repo name")
	cmd.Flags().StringVarP(&o.Path, "path", "", "", "The directory to be backed up")
	cmd.Flags().StringVarP(&o.LimitUploadRate, "limit-upload-rate", "", "", "Limits uploads to a maximum rate in KiB/s. (default: unlimited)")
	cmd.Flags().StringVarP(&o.OlaresId, "olares-id", "", "", "Olares ID")
	cmd.Flags().StringVarP(&o.CloudApiMirror, "cloud-api-mirror", "", "", "Cloud API mirror")
	cmd.Flags().StringVarP(&o.BaseDir, "base-dir", "", "", "Set Olares package base dir, defaults to $HOME/"+common.DefaultBaseDir)
}

// ~ s3
type S3BackupOption struct {
	RepoName        string
	Endpoint        string
	AccessKey       string
	SecretAccessKey string
	Path            string
	LimitUploadRate string
	OlaresId        string
	BaseDir         string
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

	cmd.Flags().StringVarP(&o.OlaresId, "olares-id", "", "", "Olares ID")
	cmd.Flags().StringVarP(&o.BaseDir, "base-dir", "", "", "Set Olares package base dir, defaults to $HOME/"+common.DefaultBaseDir)
}

// ~ cos
type CosBackupOption struct {
	RepoName        string
	Endpoint        string
	AccessKey       string
	SecretAccessKey string
	Path            string
	LimitUploadRate string
	OlaresId        string
	BaseDir         string
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

	cmd.Flags().StringVarP(&o.OlaresId, "olares-id", "", "", "Olares ID")
	cmd.Flags().StringVarP(&o.BaseDir, "base-dir", "", "", "Set Olares package base dir, defaults to $HOME/"+common.DefaultBaseDir)
}

// ~ filesystem
type FilesystemBackupOption struct {
	RepoName string
	Endpoint string
	Path     string
	OlaresId string
	BaseDir  string
}

func NewBackupFilesystemOption() *FilesystemBackupOption {
	return &FilesystemBackupOption{}
}

func (o *FilesystemBackupOption) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.RepoName, "repo-name", "", "", "Backup repo name")
	cmd.Flags().StringVarP(&o.Endpoint, "endpoint", "", "", "The endpoint of the filesystem is the local computer directory where the backup will be stored")
	cmd.Flags().StringVarP(&o.Path, "path", "", "", "The directory to be backed up")
	cmd.Flags().StringVarP(&o.OlaresId, "olares-id", "", "", "Olares ID")
	cmd.Flags().StringVarP(&o.BaseDir, "base-dir", "", "", "Set Olares package base dir, defaults to $HOME/"+common.DefaultBaseDir)
}
