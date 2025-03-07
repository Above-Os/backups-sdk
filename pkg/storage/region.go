package storage

import (
	"bytetrade.io/web3os/backups-sdk/pkg/logger"
	"bytetrade.io/web3os/backups-sdk/pkg/options"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/space"
)

type RegionOption struct {
	Space *options.SpaceRegionOptions
}

type RegionService struct {
	option *RegionOption
}

func NewRegionService(option *RegionOption) *RegionService {
	var regionService = &RegionService{
		option: option,
	}

	return regionService
}

func (r *RegionService) Regions() {
	var service Location
	if r.option != nil {
		service = &space.Space{
			OlaresDid:      r.option.Space.OlaresDid,
			AccessToken:    r.option.Space.AccessToken,
			CloudApiMirror: r.option.Space.CloudApiMirror,
		}
	}
	if err := service.Regions(); err != nil {
		logger.Errorf("Get Regions error: %v", err)
	}
}
