package storage

import (
	"context"

	"go.uber.org/zap"
	"olares.com/backups-sdk/pkg/options"
	"olares.com/backups-sdk/pkg/storage/space"
)

type RegionOption struct {
	Ctx    context.Context
	Logger *zap.SugaredLogger
	Space  *options.SpaceRegionOptions
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

func (r *RegionService) Regions() ([]map[string]string, error) {
	var service Location
	if r.option != nil {
		service = &space.Space{
			OlaresDid:      r.option.Space.OlaresDid,
			AccessToken:    r.option.Space.AccessToken,
			CloudApiMirror: r.option.Space.CloudApiMirror,
		}
	}

	return service.Regions()
}
