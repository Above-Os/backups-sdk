package util

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"olares.com/backups-sdk/pkg/constants"
	"olares.com/backups-sdk/pkg/utils"
)

func GetRestoreTargetPath(backupType string, restoreTargetPath string, uploadPath string) (string, string) {
	if backupType == constants.BackupTypeFile {
		return uploadPath, restoreTargetPath
	}

	if strings.Contains(uploadPath, "userspace/pvc-userspace") {
		return uploadPath, restoreTargetPath
	} else {
		return uploadPath, uploadPath
	}
}

func GetBackupType(tags []string) (backupType string) {
	backupType = constants.BackupTypeFile
	if tags == nil || len(tags) == 0 {
		return
	}

	for _, tag := range tags {
		e := strings.Index(tag, "=")
		if e >= 0 {
			if tag[:e] == "backup-type" {
				backupType = tag[e+1:]
				break
			}
		}
	}

	return
}

func GetMetadata(tags []string) string {
	var metadata string
	if tags == nil || len(tags) == 0 {
		return metadata
	}

	for _, tag := range tags {
		e := strings.Index(tag, "=")
		if e >= 0 {
			if tag[:e] == "metadata" {
				var tmp = tag[e+1:]
				if tmp != "" {
					b, _ := utils.Base64decode(tmp)
					if b != nil {
						metadata = string(b)
					}
					break
				}
			}
		}
	}

	return metadata
}

// files-prefix-path
func GetFilesPrefixPath(tags []string) ([]string, error) {
	var filesPrefixPath []string
	if tags == nil || len(tags) == 0 {
		return nil, fmt.Errorf("restore app backup failed, backup files path is empty")
	}

	for _, tag := range tags {
		e := strings.Index(tag, "=")
		if e >= 0 {
			if tag[:e] == "files-prefix-path" {
				var tmp = tag[e+1:]
				if tmp != "" {
					b, err := utils.Base64decode(tmp)
					if err != nil {
						return nil, fmt.Errorf("restore app backup files prefix decode error: %v", err)
					}
					if err := json.Unmarshal(b, &filesPrefixPath); err != nil {
						return nil, fmt.Errorf("restore app backup files prefix unmarshal error: %v", err)
					}
					break
				}
			}
		}
	}

	if filesPrefixPath == nil || len(filesPrefixPath) == 0 {
		return nil, fmt.Errorf("restore app backup failed, backup files path is empty")
	}

	return filesPrefixPath, nil
}

func Chmod(p string) error {
	if utils.IsExist(p) {
		return os.Chmod(p, 0755)
	}
	return utils.CreateDir(p)
}
