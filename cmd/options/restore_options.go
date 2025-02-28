package options

import (
	"bytetrade.io/web3os/backups-sdk/pkg/common"
	"github.com/spf13/cobra"
)

type SpaceRestoreOption struct {
	RepoName       string
	SnapshotId     string
	Path           string
	OlaresId       string
	BaseDir        string
	CloudApiMirror string
}

func NewRestoreSpaceOption() *SpaceRestoreOption {
	return &SpaceRestoreOption{}
}

func (o *SpaceRestoreOption) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.RepoName, "repo-name", "", "", "Backup repo name")
	cmd.Flags().StringVarP(&o.SnapshotId, "snapshot-id", "", "", "Snapshot ID")
	cmd.Flags().StringVarP(&o.Path, "path", "", "", "The directory to be restore")
	cmd.Flags().StringVarP(&o.OlaresId, "olares-id", "", "", "Olares ID")
	cmd.Flags().StringVarP(&o.CloudApiMirror, "cloud-api-mirror", "", "", "Cloud API mirror")
	cmd.Flags().StringVarP(&o.BaseDir, "base-dir", "", "", "Set Olares package base dir, defaults to $HOME/"+common.DefaultBaseDir)
}

// ~ s3
type S3RestoreOption struct {
	RepoName        string
	SnapshotId      string
	Endpoint        string
	AccessKey       string
	SecretAccessKey string
	Path            string
	OlaresId        string
	BaseDir         string
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
	cmd.Flags().StringVarP(&o.OlaresId, "olares-id", "", "", "Olares ID")
	cmd.Flags().StringVarP(&o.BaseDir, "base-dir", "", "", "Set Olares package base dir, defaults to $HOME/"+common.DefaultBaseDir)
}

// ~ cos
type CosRestoreOption struct {
	RepoName        string
	SnapshotId      string
	Endpoint        string
	AccessKey       string
	SecretAccessKey string
	Path            string
	OlaresId        string
	BaseDir         string
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

	cmd.Flags().StringVarP(&o.OlaresId, "olares-id", "", "", "Olares ID")
	cmd.Flags().StringVarP(&o.BaseDir, "base-dir", "", "", "Set Olares package base dir, defaults to $HOME/"+common.DefaultBaseDir)
}

// ~ filesystem
type FilesystemRestoreOption struct {
	RepoName   string
	SnapshotId string
	Endpoint   string
	Path       string
	OlaresId   string
	BaseDir    string
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
	cmd.Flags().StringVarP(&o.BaseDir, "base-dir", "", "", "Set Olares package base dir, defaults to $HOME/"+common.DefaultBaseDir)
}
