package adapter

import (
	"context"
	"jian-unified-system/hephaestus/hephaestus-rpc/income"
	a "jian-unified-system/hephaestus/hephaestus-rpc/internal/application"
	d "jian-unified-system/hephaestus/hephaestus-rpc/internal/domain"
	"strconv"
)

func Filter(q *income.QueryReq) d.Filter {
	if q == nil {
		return d.Filter{Limit: 50, Rate: 2000}
	}
	return d.Filter{From: q.From, To: q.To, Type: q.Type, Team: q.Team, Search: q.Search, Sort: q.Sort, Archived: q.Archived, Cursor: q.Cursor, GroupBy: q.GroupBy, Limit: q.Limit, Rate: q.RateBps}
}
func Authorize(ctx context.Context, s *a.Service, owner int64, token string) error {
	v, e := s.Session(ctx, token, "")
	if e != nil {
		return e
	}
	if v.OwnerID != owner {
		return d.Fail("FORBIDDEN", "owner mismatch")
	}
	return nil
}
func IntID(s string) (int64, error) { return d.ID(s) }
func Offset(s string) (int, error) {
	if s == "" {
		return 0, nil
	}
	v, e := strconv.Atoi(s)
	if e != nil || v < 0 || v > 10000 {
		return 0, d.Invalid("invalid row cursor")
	}
	return v, nil
}
