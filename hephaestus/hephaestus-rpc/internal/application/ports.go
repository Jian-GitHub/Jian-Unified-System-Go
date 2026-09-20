package application

import (
	"context"
	d "jian-unified-system/hephaestus/hephaestus-rpc/internal/domain"
	"jian-unified-system/hephaestus/internal/apollosso"
	"time"
)

type Revision struct {
	Version   int64     `json:"version,string"`
	Action    string    `json:"action"`
	Reason    string    `json:"reason"`
	Before    *d.Record `json:"before"`
	After     d.Record  `json:"after"`
	CreatedAt string    `json:"created_at"`
}
type Source struct {
	ID    int64  `json:"id,string"`
	Label string `json:"label"`
	Kind  string `json:"kind"`
}
type Row struct {
	Row             int64    `json:"row"`
	Data            d.Data   `json:"data"`
	SourceKey       string   `json:"source_key"`
	Fingerprint     string   `json:"fingerprint"`
	Action          string   `json:"action"`
	RecordID        int64    `json:"record_id,string"`
	ExpectedVersion int64    `json:"expected_version,string"`
	KeepOverrides   bool     `json:"keep_overrides"`
	Errors          []string `json:"errors"`
}
type Grid struct {
	Sheets   map[string][][]string `json:"sheets"`
	Date1904 bool                  `json:"date1904"`
}
type Batch struct {
	Deleted       bool     `json:"deleted"`
	Filename      string   `json:"filename"`
	Spreadsheet   string   `json:"spreadsheet"`
	Range         string   `json:"range"`
	Mapping       *Mapping `json:"mapping,omitempty"`
	ID            int64    `json:"id,string"`
	SourceID      int64    `json:"source_id,string"`
	Status        string   `json:"status"`
	Version       int64    `json:"version,string"`
	SelectionHash string   `json:"selection_hash"`
	RowCount      int64    `json:"row_count"`
	ErrorCount    int64    `json:"error_count"`
	Inserted      int64    `json:"inserted"`
	Updated       int64    `json:"updated"`
	Unchanged     int64    `json:"unchanged"`
	Excluded      int64    `json:"excluded"`
	Sheets        []string `json:"sheets"`
	CreatedAt     string   `json:"created_at"`
	ContentHash   string   `json:"content_hash"`
	Grid          Grid     `json:"grid"`
	Rows          []Row    `json:"rows"`
}
type Column struct {
	Field  string `json:"field"`
	Column int64  `json:"column"`
}
type Decision struct {
	Row             int64  `json:"row"`
	Action          string `json:"action"`
	RecordID        int64  `json:"record_id,string"`
	ExpectedVersion int64  `json:"expected_version,string"`
	KeepOverrides   bool   `json:"keep_overrides"`
}
type Mapping struct {
	RuleProfile         string     `json:"rule_profile"`
	WageColumnConfirmed bool       `json:"wage_column_confirmed"`
	Sheet               string     `json:"sheet"`
	HeaderRow           int64      `json:"header_row"`
	DateFormat          string     `json:"date_format"`
	DurationFormat      string     `json:"duration_format"`
	EmptyExpenseZero    bool       `json:"empty_expense_zero"`
	DefaultTeamSize     int64      `json:"default_team_size"`
	Columns             []Column   `json:"columns"`
	Decisions           []Decision `json:"decisions"`
}
type Session struct {
	OwnerID     int64  `json:"owner_id,string"`
	DisplayName string `json:"display_name"`
	CSRFToken   string `json:"csrf_token"`
	ExpiresAt   string `json:"expires_at"`
}
type Tx interface {
	Record(id int64) (d.Record, error)
	BySource(source int64, key string) (*d.Record, error)
	Candidates(data d.Data) ([]d.Record, error)
	SaveRecord(record *d.Record, before *d.Record, action, reason string) error
	Batch(id int64) (Batch, error)
	// BatchByHash returns only an editable batch; terminal snapshots are history.
	BatchByHash(source int64, hash string) (*Batch, error)
	SaveBatch(batch *Batch) error
	Source(id int64) (Source, error)
	SaveSource(label string) (Source, error)
	DeleteSource(id int64) error
	Memo(operation, key, hash string) ([]byte, error)
	Remember(operation, key, hash string, result []byte) error
}
type Repository interface {
	Within(context.Context, int64, func(Tx) error) error
	Records(context.Context, int64, d.Filter, int, int) ([]d.Record, error)
	Revisions(context.Context, int64, int64) ([]Revision, error)
	Sources(context.Context, int64) ([]Source, error)
	Batches(context.Context, int64) ([]Batch, error)
	ResolveOwner(context.Context, string, string) (int64, error)
	Health(context.Context) error
}
type Parser interface {
	Parse(context.Context, int64, string, string) (Grid, string, error)
}
type Authority interface {
	Introspect(context.Context, string) (apollosso.Session, error)
	Revoke(context.Context, string) error
}
type Service struct {
	Authority Authority
	Repo      Repository
	Parser    Parser
	Now       func() time.Time
}
