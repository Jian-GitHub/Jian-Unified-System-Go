package application

import (
	"context"
	"fmt"
	d "jian-unified-system/hephaestus/hephaestus-rpc/internal/domain"
	"math/big"
	"strconv"
	"strings"
	"time"
)

func normalizedDate(s, format string, date1904 bool) (string, error) {
	if format == "excel" {
		s = strings.TrimSuffix(s, ".0")
		n, e := strconv.Atoi(s)
		if e != nil || n < 0 || n > 2958465 || (!date1904 && n == 60) {
			return "", d.Invalid("invalid Excel date serial")
		}
		base := time.Date(1899, 12, 31, 0, 0, 0, 0, time.UTC)
		if date1904 {
			base = time.Date(1904, 1, 1, 0, 0, 0, 0, time.UTC)
		} else if n > 60 {
			n--
		}
		result := base.AddDate(0, 0, n).Format("2006-01-02")
		_, e = d.Date(result)
		return result, e
	}
	layout := "2006-01-02"
	switch format {
	case "iso":
	case "dmy":
		layout = "02/01/2006"
	case "mdy":
		layout = "01/02/2006"
	default:
		return "", d.Invalid("date_format must be iso, dmy, mdy or excel")
	}
	t, e := time.Parse(layout, s)
	if e != nil || t.Format(layout) != s {
		return "", d.Invalid("date does not match the selected format")
	}
	result := t.Format("2006-01-02")
	_, e = d.Date(result)
	return result, e
}
func duration(s, format string) (int64, error) {
	if s == "" {
		return 0, nil
	}
	switch format {
	case "minutes":
		n, e := strconv.ParseInt(s, 10, 64)
		if e != nil {
			return 0, d.Invalid("invalid minutes")
		}
		return n, nil
	case "hours":
		if strings.ContainsAny(s, "eE/,+-") {
			return 0, d.Invalid("invalid decimal hours")
		}
		v, ok := new(big.Rat).SetString(s)
		if !ok {
			return 0, d.Invalid("invalid hours")
		}
		v.Mul(v, big.NewRat(60, 1))
		if !v.IsInt() || !v.Num().IsInt64() {
			return 0, d.Invalid("hours must represent whole minutes")
		}
		return v.Num().Int64(), nil
	case "hh:mm":
		p := strings.Split(s, ":")
		if len(p) != 2 {
			return 0, d.Invalid("expected HH:mm")
		}
		h, e := strconv.ParseInt(p[0], 10, 64)
		m, e2 := strconv.ParseInt(p[1], 10, 64)
		if e != nil || e2 != nil || h < 0 || h > 24 || m < 0 || m > 59 {
			return 0, d.Invalid("invalid HH:mm")
		}
		return h*60 + m, nil
	}
	return 0, d.Invalid("duration_format must be minutes, hours or hh:mm")
}
func validateMapping(m Mapping) error {
	if m.RuleProfile != "" && m.RuleProfile != "aimerhq-v1" {
		return d.Invalid("unknown rule profile")
	}
	if m.RuleProfile == "aimerhq-v1" && !m.WageColumnConfirmed {
		return d.Invalid("confirm the actual company wage column; this workbook has conflicting Net Wage / Payment Status headers")
	}
	if m.HeaderRow < 1 || m.HeaderRow > 100 || m.DefaultTeamSize < 1 || m.DefaultTeamSize > 1000 {
		return d.Invalid("invalid header row or default team size")
	}
	if _, e := normalizedDate(map[string]string{"iso": "2026-01-01", "dmy": "01/01/2026", "mdy": "01/01/2026", "excel": "46023"}[m.DateFormat], m.DateFormat, false); e != nil {
		return e
	}
	if m.DurationFormat != "minutes" && m.DurationFormat != "hours" && m.DurationFormat != "hh:mm" {
		return d.Invalid("invalid duration format")
	}
	seen := map[string]bool{}
	for _, c := range m.Columns {
		if seen[c.Field] || c.Column < 0 || c.Column > 99 {
			return d.Invalid("duplicate mapping or invalid zero-based column")
		}
		seen[c.Field] = true
		switch c.Field {
		case "service_date", "customer", "job_type", "detail", "note", "pricing_category", "duration", "team_size", "gross", "expense", "source_key", "payment_status":
		default:
			return d.Invalid("unknown mapping field")
		}
	}
	if !seen["service_date"] || !seen["gross"] {
		return d.Invalid("service_date and gross columns must be mapped")
	}
	if m.RuleProfile == "aimerhq-v1" {
		for _, field := range []string{"customer", "job_type", "detail", "duration", "note", "expense"} {
			if !seen[field] {
				return d.Invalid("AimerHQ mapping requires customer, job type, detail, duration, note and Cost")
			}
		}
		used := map[int64]bool{}
		for _, column := range m.Columns {
			if used[column.Column] {
				return d.Invalid("AimerHQ fields must use separate columns")
			}
			used[column.Column] = true
		}
	}
	return nil
}
func normalize(cells []string, m Mapping, date1904 bool) (d.Data, string, error) {
	values := map[string]string{}
	for _, c := range m.Columns {
		if int(c.Column) < len(cells) {
			values[c.Field] = strings.TrimSpace(cells[c.Column])
		}
	}
	data := d.Data{Customer: values["customer"], JobType: values["job_type"], Detail: values["detail"], Note: values["note"], PricingCategory: values["pricing_category"], TeamSize: m.DefaultTeamSize}
	data.PaymentStatus = values["payment_status"]
	var e error
	data.ServiceDate, e = normalizedDate(values["service_date"], m.DateFormat, date1904)
	if m.RuleProfile == "aimerhq-v1" {
		data.ServiceDate, e = aimerDate(values["service_date"], m.DateFormat, date1904)
	}
	if e != nil {
		return data, "", e
	}
	data.DurationMinutes, e = duration(values["duration"], m.DurationFormat)
	if e != nil {
		return data, "", e
	}
	if v := values["team_size"]; v != "" {
		data.TeamSize, e = parsePositive(v)
		if e != nil {
			return data, "", e
		}
	}
	data.Gross, e = d.DecimalCents(values["gross"])
	if m.RuleProfile == "aimerhq-v1" {
		data.Gross, e = aimerAmount(values["gross"])
	}
	if e != nil {
		return data, "", e
	}
	expense := values["expense"]
	if expense == "" && (m.EmptyExpenseZero || m.RuleProfile == "aimerhq-v1") {
		expense = "0"
	}
	data.Expense, e = d.DecimalCents(expense)
	if m.RuleProfile == "aimerhq-v1" {
		data.Expense, e = aimerAmount(expense)
	}
	if e != nil {
		return data, "", e
	}
	key := values["source_key"]
	if m.RuleProfile == "aimerhq-v1" {
		data, e = d.ApplyAimerHQ(data)
		if e != nil {
			return data, "", e
		}
		if key == "" {
			key = "aimer:" + d.Hash([]string{data.ServiceDate, strings.Join(strings.Fields(strings.ToLower(data.Customer)), " "), strings.Join(strings.Fields(strings.ToLower(data.JobType)), " ")})
		}
	}
	if len(key) > 100 {
		return data, "", d.Invalid("source key exceeds 100 bytes")
	}
	return data, key, data.Validate()
}
func (s *Service) Preview(ctx context.Context, owner, id, version int64, key string, m Mapping) (Batch, error) {
	if e := validateMapping(m); e != nil {
		return Batch{}, e
	}
	return command(ctx, s, owner, "import.preview", key, []any{id, version, m}, false, func(tx Tx) (Batch, error) {
		b, e := tx.Batch(id)
		if e != nil {
			return b, e
		}
		if e = d.Version(b.Version, version); e != nil {
			return b, e
		}
		if b.Status == "committed" || b.Status == "cancelled" || b.Deleted {
			return b, d.Fail("CONFLICT", "batch is terminal")
		}
		grid, ok := b.Grid.Sheets[m.Sheet]
		if !ok || int(m.HeaderRow) > len(grid) {
			return b, d.Invalid("unknown sheet or header row")
		}
		decisions := map[int64]Decision{}
		for _, v := range m.Decisions {
			if _, ok := decisions[v.Row]; ok {
				return b, d.Invalid("duplicate row decision")
			}
			if v.Row <= m.HeaderRow || v.Row > int64(len(grid)) {
				return b, d.Invalid("decision row is outside data range")
			}
			if v.Action != "exclude" && v.Action != "new" && v.Action != "update" && v.Action != "unchanged" {
				return b, d.Invalid("unsupported row action")
			}
			decisions[v.Row] = v
		}
		b.Rows = []Row{}
		b.ErrorCount = 0
		b.Excluded = 0
		b.Inserted = 0
		b.Updated = 0
		b.Unchanged = 0
		keys := map[string]int{}
		targets := map[int64]bool{}
		for i := int(m.HeaderRow); i < len(grid); i++ {
			if e := ctx.Err(); e != nil {
				return b, e
			}
			row := Row{Row: int64(i + 1), Action: "insert", Errors: []string{}}
			dec, explicit := decisions[row.Row]
			if m.RuleProfile == "aimerhq-v1" && aimerNonWorkRow(grid[i], m) {
				dec.Action = "exclude"
				explicit = true
			}
			if explicit && dec.Action == "exclude" {
				row.Action = "exclude"
				b.Excluded++
				b.Rows = append(b.Rows, row)
				continue
			}
			data, sourceKey, err := normalize(grid[i], m, b.Grid.Date1904)
			row.Data = data
			row.Fingerprint = d.Hash(data)
			row.SourceKey = "fp:" + row.Fingerprint
			if sourceKey != "" {
				row.SourceKey = "key:" + sourceKey
			}
			if explicit && dec.Action == "new" && sourceKey == "" {
				row.SourceKey = fmt.Sprintf("row:%d:%d", id, row.Row)
			}
			if err != nil {
				row.Errors = append(row.Errors, err.Error())
			} else {
				existing, err := tx.BySource(b.SourceID, row.SourceKey)
				if err != nil {
					return b, err
				}
				if explicit && (dec.Action == "update" || dec.Action == "unchanged") {
					r, err := tx.Record(dec.RecordID)
					if err != nil {
						return b, err
					}
					if r.SourceID != b.SourceID {
						return b, d.Invalid("target must belong to the same source")
					}
					if sourceKey != "" && r.SourceKey != row.SourceKey {
						return b, d.Invalid("stable source key cannot be reassigned")
					}
					if existing != nil && existing.ID != r.ID {
						return b, d.Invalid("source key belongs to a different record")
					}
					existing = &r
					row.SourceKey = r.SourceKey
					if err = d.Version(r.Version, dec.ExpectedVersion); err != nil {
						return b, err
					}
				}
				if existing != nil {
					row.RecordID = existing.ID
					row.ExpectedVersion = existing.Version
					row.Action = "update"
					if existing.Archived {
						row.Errors = append(row.Errors, "matched record is archived; restore separately or exclude")
					}
					if sourceKey == "" && !explicit {
						row.Errors = append(row.Errors, "fingerprint is only a candidate; choose update, unchanged, new or exclude")
					}
					if explicit && dec.Action == "new" {
						row.Errors = append(row.Errors, "stable key already exists; new is not allowed")
					}
					if len(existing.Overrides) > 0 && !explicit {
						row.Errors = append(row.Errors, "manual overrides exist; explicitly keep or adopt source values")
					}
					row.KeepOverrides = dec.KeepOverrides
					if d.Hash(existing.Base) == row.Fingerprint && (!explicit || dec.Action != "update" || dec.KeepOverrides || len(existing.Overrides) == 0) {
						row.Action = "unchanged"
						row.KeepOverrides = true
					}
					if explicit && dec.Action == "unchanged" {
						if d.Hash(existing.Base) != row.Fingerprint {
							row.Errors = append(row.Errors, "source differs; choose update or exclude")
						}
						row.Action = "unchanged"
						row.KeepOverrides = true
					}
					if targets[existing.ID] {
						return b, d.Invalid("multiple input rows target the same record")
					}
					targets[existing.ID] = true
				} else if sourceKey == "" && !(explicit && dec.Action == "new") {
					candidates, err := tx.Candidates(data)
					if err != nil {
						return b, err
					}
					if len(candidates) > 0 {
						row.Errors = append(row.Errors, "possible existing work on the same date; choose new, update or exclude")
					}
				}
			}
			if previous, ok := keys[row.SourceKey]; ok {
				row.Errors = append(row.Errors, "duplicate source identity inside this batch")
				b.Rows[previous].Errors = append(b.Rows[previous].Errors, "duplicate source identity inside this batch")
			} else {
				keys[row.SourceKey] = len(b.Rows)
			}
			b.Rows = append(b.Rows, row)
		}
		for _, r := range b.Rows {
			if len(r.Errors) > 0 {
				b.ErrorCount++
			}
		}
		b.RowCount = int64(len(b.Rows))
		b.Status = "awaiting_review"
		b.Mapping = &m
		b.Version++
		b.SelectionHash = d.Hash(b.Rows)
		e = tx.SaveBatch(&b)
		return b, e
	})
}
func (s *Service) Commit(ctx context.Context, owner, id, version int64, key, selection string) (Batch, error) {
	return command(ctx, s, owner, "import.commit", key, []any{id, version, selection}, true, func(tx Tx) (Batch, error) {
		b, e := tx.Batch(id)
		if e != nil {
			return b, e
		}
		if b.Status == "committed" {
			if b.SelectionHash == selection {
				return b, nil
			}
			return b, d.Fail("CONFLICT", "batch already committed")
		}
		if e = d.Version(b.Version, version); e != nil {
			return b, e
		}
		if b.Deleted || b.Status != "awaiting_review" || b.SelectionHash != selection || selection == "" {
			return b, d.Conflict()
		}
		if _, e := tx.Source(b.SourceID); e != nil {
			return b, e
		}
		if b.ErrorCount > 0 {
			return b, d.Invalid("resolve or exclude all invalid rows")
		}
		for _, row := range b.Rows {
			if e := ctx.Err(); e != nil {
				return b, e
			}
			if row.Action == "exclude" {
				continue
			}
			if len(row.Errors) > 0 {
				return b, d.Invalid("preview contains errors")
			}
			if row.RecordID != 0 {
				r, e := tx.Record(row.RecordID)
				if e != nil {
					return b, e
				}
				if e = d.Version(r.Version, row.ExpectedVersion); e != nil {
					return b, e
				}
				if r.Archived {
					return b, d.Conflict()
				}
				if row.Action == "unchanged" {
					b.Unchanged++
					continue
				}
				before := clone(r)
				r.Base = row.Data
				if !row.KeepOverrides {
					r.Overrides = []d.Change{}
				}
				if e = tx.SaveRecord(&r, &before, "source-updated", fmt.Sprintf("import batch %d", id)); e != nil {
					return b, e
				}
				b.Updated++
			} else {
				existing, e := tx.BySource(b.SourceID, row.SourceKey)
				if e != nil {
					return b, e
				}
				if existing != nil {
					return b, d.Conflict()
				}
				r := d.Record{Base: row.Data, SourceID: b.SourceID, SourceKey: row.SourceKey}
				if e = tx.SaveRecord(&r, nil, "imported", fmt.Sprintf("import batch %d", id)); e != nil {
					return b, e
				}
				b.Inserted++
			}
		}
		b.Status = "committed"
		b.Version++
		e = tx.SaveBatch(&b)
		return b, e
	})
}
