package application

import (
	"context"
	"errors"
	d "jian-unified-system/hephaestus/hephaestus-rpc/internal/domain"
	"jian-unified-system/hephaestus/internal/apollosso"
)

type GoogleReader interface {
	GoogleSheets(context.Context, string, string, string, string, string, string) (apollosso.GoogleResult, error)
}

func (s *Service) StageGoogle(ctx context.Context, owner, source int64, key, token, spreadsheet, area string) (Batch, error) {
	if e := keyValid(key, true); e != nil {
		return Batch{}, e
	}
	if e := s.Repo.Within(ctx, owner, func(tx Tx) error { _, e := tx.Source(source); return e }); e != nil {
		return Batch{}, e
	}
	reader, ok := s.Authority.(GoogleReader)
	if !ok {
		return Batch{}, apollosso.ErrUnavailable
	}
	data, e := reader.GoogleSheets(ctx, token, "read", "", "", spreadsheet, area)
	if errors.Is(e, apollosso.ErrGoogleAuthorization) {
		return Batch{}, d.Fail("CONFLICT", "Google Sheets authorization required; connect again")
	}
	if errors.Is(e, apollosso.ErrGoogleInput) {
		return Batch{}, d.Invalid("check spreadsheet link, range and size")
	}
	if errors.Is(e, apollosso.ErrGoogleAccess) {
		return Batch{}, d.Invalid("Google spreadsheet is not shared with the linked Google account")
	}
	if e != nil {
		return Batch{}, authorityError(e)
	}
	if len(data.Rows) == 0 || len(data.Rows) > 10000 {
		return Batch{}, d.Invalid("Google range must contain 1..10000 rows")
	}
	for _, r := range data.Rows {
		if len(r) > 100 {
			return Batch{}, d.Invalid("Google range exceeds 100 columns")
		}
	}
	// Google supplies formatted cell text; mapping explicitly chooses date and amount formats.
	grid := Grid{Sheets: map[string][][]string{"Google Sheets": data.Rows}}
	hash := d.Hash([]any{"google-sheets", spreadsheet, area, grid})
	return s.stageGrid(ctx, owner, source, key, grid, hash, Batch{Spreadsheet: spreadsheet, Range: area})
}
