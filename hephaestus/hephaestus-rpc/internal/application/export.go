package application

import (
	"bytes"
	"context"
	"encoding/csv"
	d "jian-unified-system/hephaestus/hephaestus-rpc/internal/domain"
	"strconv"
	"strings"
)

func csvText(s string) string {
	v := strings.TrimLeft(s, " \t\r\n")
	if len(v) > 0 && strings.ContainsRune("=+-@", rune(v[0])) {
		return "'" + s
	}
	return s
}
func (s *Service) Export(ctx context.Context, owner int64, f d.Filter) (string, error) {
	rows, e := s.All(ctx, owner, f)
	if e != nil {
		return "", e
	}
	var b bytes.Buffer
	w := csv.NewWriter(&b)
	if e = w.Write([]string{"id", "service_date", "customer", "job_type", "detail", "note", "duration_minutes", "team_size", "gross_nzd", "expense_nzd", "tax_estimate_nzd", "cash_estimate_nzd", "wage_status"}); e != nil {
		return "", e
	}
	for _, r := range rows {
		if e = ctx.Err(); e != nil {
			return "", e
		}
		total, e := d.Total([]d.Record{r}, f.Rate)
		if e != nil {
			return "", e
		}
		tax, cash := "", ""
		if r.Data.Gross != nil {
			tax = d.FormatMoney(&total.Tax)
			if r.Data.Expense != nil {
				cash = d.FormatMoney(&total.Cash)
			}
		}
		v := r.Data
		if e = w.Write([]string{strconv.FormatInt(r.ID, 10), v.ServiceDate, csvText(v.Customer), csvText(v.JobType), csvText(v.Detail), csvText(v.Note), strconv.FormatInt(v.DurationMinutes, 10), strconv.FormatInt(v.TeamSize, 10), d.FormatMoney(v.Gross), d.FormatMoney(v.Expense), tax, cash, r.WageStatus}); e != nil {
			return "", e
		}
		if b.Len() > 16<<20 {
			return "", d.Fail("TOO_LARGE", "export exceeds 16 MiB; narrow filters")
		}
	}
	w.Flush()
	return b.String(), w.Error()
}
