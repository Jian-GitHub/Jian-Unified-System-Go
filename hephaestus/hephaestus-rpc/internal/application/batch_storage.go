package application

import (
	"context"
	d "jian-unified-system/hephaestus/hephaestus-rpc/internal/domain"
	"strconv"
)

// Removing batch history never deletes or rolls back imported business records.
func (s *Service) TrashBatch(ctx context.Context, owner, id, version int64, key string, deleted bool) (Batch, error) {
	return command(ctx, s, owner, "import.trash", key, []any{id, version, deleted}, true, func(tx Tx) (Batch, error) {
		b, e := tx.Batch(id)
		if e != nil {
			return b, e
		}
		if e = d.Version(b.Version, version); e != nil {
			return b, e
		}
		b.Deleted = deleted
		b.Version++
		e = tx.SaveBatch(&b)
		return b, e
	})
}

type SheetRow struct {
	Row   int64    `json:"row"`
	Cells []string `json:"cells"`
}
type SheetData struct {
	Sheet      string     `json:"sheet"`
	Items      []SheetRow `json:"items"`
	NextCursor string     `json:"next_cursor"`
	Total      int64      `json:"total"`
}

func (s *Service) ReadSheet(ctx context.Context, owner, id int64, sheet, cursor string) (SheetData, error) {
	b, e := s.GetBatch(ctx, owner, id)
	if e != nil {
		return SheetData{}, e
	}
	if sheet == "" && len(b.Sheets) > 0 {
		sheet = b.Sheets[0]
	}
	grid, ok := b.Grid.Sheets[sheet]
	if !ok {
		return SheetData{}, d.Invalid("unknown sheet")
	}
	start := 0
	if cursor != "" {
		start, e = strconv.Atoi(cursor)
		if e != nil || start < 0 || start > len(grid) {
			return SheetData{}, d.Invalid("invalid sheet cursor")
		}
	}
	end := min(start+100, len(grid))
	result := SheetData{Sheet: sheet, Total: int64(len(grid)), Items: []SheetRow{}}
	for i := start; i < end; i++ {
		result.Items = append(result.Items, SheetRow{Row: int64(i + 1), Cells: grid[i]})
	}
	if end < len(grid) {
		result.NextCursor = strconv.Itoa(end)
	}
	return result, nil
}
