package application

import (
	"context"
	d "jian-unified-system/hephaestus/hephaestus-rpc/internal/domain"
	"reflect"
	"testing"
	"time"
)

// JSON copies simulate the persisted snapshot boundary; unexpected record
// writes panic through the embedded interface instead of silently succeeding.
type batchRepo struct {
	Repository
	tx    *batchTx
	owner int64
}

func (r *batchRepo) Within(_ context.Context, owner int64, f func(Tx) error) error {
	if owner != r.owner {
		return d.Fail("NOT_FOUND", "not found")
	}
	return f(r.tx)
}

type batchTx struct {
	Tx
	batch Batch
	saves int
	memos map[string][]byte
}

func (tx *batchTx) Batch(id int64) (Batch, error) {
	if id != tx.batch.ID {
		return Batch{}, d.Fail("NOT_FOUND", "not found")
	}
	return clone(tx.batch), nil
}
func (tx *batchTx) SaveBatch(b *Batch) error {
	if b.ID == 0 {
		b.ID = 10
	}
	tx.batch = clone(*b)
	tx.saves++
	return nil
}
func (tx *batchTx) Source(id int64) (Source, error)               { return Source{ID: id}, nil }
func (tx *batchTx) BatchByHash(_ int64, _ string) (*Batch, error) { return nil, nil }
func (tx *batchTx) BySource(_ int64, _ string) (*d.Record, error) { return nil, nil }
func (tx *batchTx) Candidates(_ d.Data) ([]d.Record, error)       { return nil, nil }
func (tx *batchTx) Memo(op, key, hash string) ([]byte, error)     { return tx.memos[op+key+hash], nil }
func (tx *batchTx) Remember(op, key, hash string, b []byte) error {
	tx.memos[op+key+hash] = b
	return nil
}
func newBatchService() (*Service, *batchTx) {
	tx := &batchTx{memos: map[string][]byte{}}
	return &Service{Repo: &batchRepo{tx: tx, owner: 1}, Now: func() time.Time { return time.Date(2026, 9, 17, 0, 0, 0, 0, time.UTC) }}, tx
}

func TestSavedImportMappingAndSheetSnapshot(t *testing.T) {
	s, tx := newBatchService()
	ctx := context.Background()
	grid := Grid{Sheets: map[string][][]string{"CSV": {{"Date", "Wage"}, {"2026-09-17", "123.45"}, {"invalid", ""}}}}
	b, e := s.stageGrid(ctx, 1, 7, "stage-unique-key-1", grid, "hash", Batch{Filename: "work.csv", Spreadsheet: "saved-reference", Range: "CSV!A1:B3"})
	if e != nil {
		t.Fatal(e)
	}
	m := Mapping{Sheet: "CSV", HeaderRow: 1, DateFormat: "iso", DurationFormat: "minutes", DefaultTeamSize: 1, Columns: []Column{{"service_date", 0}, {"gross", 1}}, Decisions: []Decision{{Row: 3, Action: "exclude"}}}
	b, e = s.Preview(ctx, 1, b.ID, b.Version, "preview-unique-key", m)
	if e != nil || b.ErrorCount != 0 || b.Excluded != 1 {
		t.Fatalf("preview: %+v %v", b, e)
	}
	saved, e := s.GetBatch(ctx, 1, b.ID)
	if e != nil || !reflect.DeepEqual(saved.Mapping, &m) || saved.Filename != "work.csv" || saved.Spreadsheet != "saved-reference" || saved.Range != "CSV!A1:B3" {
		t.Fatalf("snapshot lost metadata: %+v %v", saved, e)
	}
	sheet, e := s.ReadSheet(ctx, 1, b.ID, "", "")
	if e != nil || sheet.Total != 3 || !reflect.DeepEqual(sheet.Items[2].Cells, grid.Sheets["CSV"][2]) {
		t.Fatalf("raw excluded cells must remain available: %+v %v", sheet, e)
	}
	if _, e = s.ReadSheet(ctx, 2, b.ID, "", ""); e == nil {
		t.Fatal("owner boundary bypassed")
	}
	if tx.saves != 2 {
		t.Fatal("read mutated saved batch")
	}
}

func TestSheetPaginationAndValidation(t *testing.T) {
	s, tx := newBatchService()
	ctx := context.Background()
	rows := make([][]string, 205)
	for i := range rows {
		rows[i] = []string{"cell"}
	}
	tx.batch = Batch{ID: 10, Sheets: []string{"Sheet"}, Grid: Grid{Sheets: map[string][][]string{"Sheet": rows}}}
	for _, tc := range []struct {
		cursor string
		count  int
		next   string
		first  int64
	}{{"", 100, "100", 1}, {"100", 100, "200", 101}, {"200", 5, "", 201}, {"205", 0, "", 0}} {
		got, e := s.ReadSheet(ctx, 1, 10, "", tc.cursor)
		if e != nil || len(got.Items) != tc.count || got.NextCursor != tc.next || got.Total != 205 {
			t.Fatalf("page %+v: %+v %v", tc, got, e)
		}
		if tc.count > 0 && got.Items[0].Row != tc.first {
			t.Fatal("row numbers reset across pages")
		}
	}
	for _, cursor := range []string{"-1", "206", "abc", "99999999999999999999999"} {
		if _, e := s.ReadSheet(ctx, 1, 10, "", cursor); e == nil {
			t.Fatalf("invalid cursor accepted: %s", cursor)
		}
	}
	if _, e := s.ReadSheet(ctx, 1, 10, "Missing", ""); e == nil {
		t.Fatal("unknown sheet accepted")
	}
}

func TestBatchTrashRestoreVersionsAndNoRecordMutation(t *testing.T) {
	s, tx := newBatchService()
	ctx := context.Background()
	tx.batch = Batch{ID: 10, Version: 3, Status: "committed", SelectionHash: "selection", Rows: []Row{{Row: 2, RecordID: 20}}}
	if _, e := s.TrashBatch(ctx, 1, 10, 2, "trash-stale-key-1", true); e == nil || tx.saves != 0 {
		t.Fatal("stale delete changed batch")
	}
	b, e := s.TrashBatch(ctx, 1, 10, 3, "trash-current-key", true)
	if e != nil || !b.Deleted || b.Version != 4 || b.Status != "committed" || len(tx.batch.Rows) != 1 {
		t.Fatalf("trash altered import history: %+v %v", b, e)
	}
	if _, e = s.TrashBatch(ctx, 1, 10, 3, "trash-current-key", true); e != nil || tx.saves != 1 {
		t.Fatal("retry was not idempotent", e)
	}
	if _, e = s.Cancel(ctx, 1, 10, 4, "cancel-current-key"); e == nil {
		t.Fatal("deleted batch cancelled")
	}
	b, e = s.TrashBatch(ctx, 1, 10, 4, "restore-current-key", false)
	if e != nil || b.Deleted || b.Version != 5 || tx.saves != 2 {
		t.Fatalf("restore: %+v %v", b, e)
	}
}
