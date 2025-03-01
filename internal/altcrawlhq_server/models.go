package altcrawlhqserver

import (
	"github.com/internetarchive/gocrawlhq"
	"github.com/saveweb/altcrawlhq_server/internal/sqlc_model"
)

func SqlcURL2hqURL(in *sqlc_model.Url) *gocrawlhq.URL {
	return &gocrawlhq.URL{
		ID:        in.ID,
		Value:     in.Value,
		Via:       in.Via,
		Host:      in.Host,
		Path:      in.Path,
		Type:      in.Type,
		Crawler:   in.Crawler,
		Status:    in.Status,
		LiftOff:   in.LiftOff,
		Timestamp: in.Timestamp,
	}
}

func URL2SqlcURL(in *gocrawlhq.URL, project string) *sqlc_model.Url {
	return &sqlc_model.Url{
		Project: project,

		ID:        in.ID,
		Value:     in.Value,
		Via:       in.Via,
		Host:      in.Host,
		Path:      in.Path,
		Type:      in.Type,
		Crawler:   in.Crawler,
		Status:    in.Status,
		LiftOff:   in.LiftOff,
		Timestamp: in.Timestamp,
	}
}

func ToCreateURLParams(in *sqlc_model.Url) sqlc_model.CreateURLParams {
	return sqlc_model.CreateURLParams{
		Project: in.Project,

		ID:        in.ID,
		Value:     in.Value,
		Via:       in.Via,
		Host:      in.Host,
		Path:      in.Path,
		Type:      in.Type,
		Crawler:   in.Crawler,
		Status:    in.Status,
		LiftOff:   in.LiftOff,
		Timestamp: in.Timestamp,
	}
}
