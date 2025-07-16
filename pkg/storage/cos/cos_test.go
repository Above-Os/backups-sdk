package cos

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestA(t *testing.T) {
	TestCosHostFormatA(t)
	TestCosHostFormatB(t)

}

func TestCosHostFormatA(t *testing.T) {
	//
	var c = &TencentCloud{
		RepoId:   "00000000-0000-0000-0000-000000000000",
		RepoName: "mybackup",
		Endpoint: "https://mytest-bucket.cos.ap-tokyo.myqcloud.com",
	}
	repo, err := c.FormatRepository()
	assert.Equal(t, err, nil)
	assert.Equal(t, repo.Url, "s3:https://cos.ap-tokyo.myqcloud.com/mytest-bucket/olares-backups/mybackup-00000000-0000-0000-0000-000000000000")

	//
	c = &TencentCloud{
		RepoId:   "00000000-0000-0000-0000-000000000000",
		RepoName: "mybackup",
		Endpoint: "https://mytest-bucket.cos.ap-tokyo.myqcloud.com/",
	}
	repo, err = c.FormatRepository()
	assert.Equal(t, err, nil)
	assert.Equal(t, repo.Url, "s3:https://cos.ap-tokyo.myqcloud.com/mytest-bucket/olares-backups/mybackup-00000000-0000-0000-0000-000000000000")

	//
	c = &TencentCloud{
		RepoId:   "00000000-0000-0000-0000-000000000000",
		RepoName: "mybackup",
		Endpoint: "https://mytest-bucket.cos.ap-tokyo.myqcloud.com/folder1/",
	}
	repo, err = c.FormatRepository()
	assert.Equal(t, err, nil)
	assert.Equal(t, repo.Url, "s3:https://cos.ap-tokyo.myqcloud.com/mytest-bucket/folder1/olares-backups/mybackup-00000000-0000-0000-0000-000000000000")

	//

}

func TestCosHostFormatB(t *testing.T) {
	c := &TencentCloud{
		RepoId:   "00000000-0000-0000-0000-000000000000",
		RepoName: "mybackup",
		Endpoint: "https://cos.ap-tokyo.myqcloud.com/mytest-bucket",
	}
	repo, err := c.FormatRepository()
	assert.Equal(t, err, nil)
	assert.Equal(t, repo.Url, "s3:https://cos.ap-tokyo.myqcloud.com/mytest-bucket/olares-backups/mybackup-00000000-0000-0000-0000-000000000000")

	//
	c = &TencentCloud{
		RepoId:   "00000000-0000-0000-0000-000000000000",
		RepoName: "mybackup",
		Endpoint: "https://cos.ap-tokyo.myqcloud.com/mytest-bucket/",
	}
	repo, err = c.FormatRepository()
	assert.Equal(t, err, nil)
	assert.Equal(t, repo.Url, "s3:https://cos.ap-tokyo.myqcloud.com/mytest-bucket/olares-backups/mybackup-00000000-0000-0000-0000-000000000000")

	//
	c = &TencentCloud{
		RepoId:   "00000000-0000-0000-0000-000000000000",
		RepoName: "mybackup",
		Endpoint: "https://cos.ap-tokyo.myqcloud.com/mytest-bucket/folder",
	}
	repo, err = c.FormatRepository()
	assert.Equal(t, err, nil)
	assert.Equal(t, repo.Url, "s3:https://cos.ap-tokyo.myqcloud.com/mytest-bucket/folder/olares-backups/mybackup-00000000-0000-0000-0000-000000000000")

	//
	c = &TencentCloud{
		RepoId:   "00000000-0000-0000-0000-000000000000",
		RepoName: "mybackup",
		Endpoint: "https://cos.ap-tokyo.myqcloud.com/mytest-bucket/folder/",
	}
	repo, err = c.FormatRepository()
	assert.Equal(t, err, nil)
	assert.Equal(t, repo.Url, "s3:https://cos.ap-tokyo.myqcloud.com/mytest-bucket/folder/olares-backups/mybackup-00000000-0000-0000-0000-000000000000")
}
