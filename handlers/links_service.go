package handlers

import (
	db "code/db/generated"
	"code/utils"
	"context"
	"database/sql"
	"errors"
)

type LinkService struct {
	Querier db.Querier
}

func NewLinkService(q db.Querier) *LinkService {
	return &LinkService{Querier: q}
}

func (s *LinkService) Create(ctx context.Context, body CreateLinkDTO) (db.Link, error) {
	params := db.CreateLinkParams{
		OriginalUrl: body.OriginalUrl,
	}

	if body.ShortName != nil && *body.ShortName != "" {
		params.ShortName = *body.ShortName
		link, err := s.Querier.CreateLink(ctx, params)
		if err != nil {
			return db.Link{}, handleDBError(err)
		}
		return link, nil
	}

	for attempts := 0; attempts < 5; attempts++ {
		shortName, err := utils.RandomString(6)

		if err != nil {
			return db.Link{}, err
		}
		params.ShortName = shortName
		link, err := s.Querier.CreateLink(ctx, params)
		if err == nil {
			return link, nil
		}
		dbErr := handleDBError(err)
		if errors.Is(dbErr, ErrorShortNameExists) {
			continue
		}
		return db.Link{}, dbErr
	}
	return db.Link{}, errors.New("failed to generate unique short name")
}

func (s *LinkService) Update(ctx context.Context, id int64, body UpdateLinkDTO) (db.Link, error) {
	params := db.UpdateLinkParams{ID: id}
	if body.OriginalUrl != nil {
		params.OriginalUrl = sql.NullString{String: *body.OriginalUrl, Valid: true}
	}
	if body.ShortName != nil {
		params.ShortName = sql.NullString{String: *body.ShortName, Valid: true}
	}
	link, err := s.Querier.UpdateLink(ctx, params)
	if err == nil {
		return link, nil
	}
	return db.Link{}, handleDBError(err)
}
