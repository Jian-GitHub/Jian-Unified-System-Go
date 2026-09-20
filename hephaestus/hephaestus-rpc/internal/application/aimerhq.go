package application

import (
	d "jian-unified-system/hephaestus/hephaestus-rpc/internal/domain"
	"regexp"
	"strings"
	"time"
)

var aimerMoney = regexp.MustCompile(`^(?:[0-9]+|[0-9]{1,3}(?:,[0-9]{3})+)(?:\.[0-9]{1,2})?$`)

func aimerAmount(s string) (*int64, error) {
	s = strings.TrimSpace(s)
	for _, prefix := range []string{"NZD", "NZ$", "$"} {
		s = strings.TrimSpace(strings.TrimPrefix(s, prefix))
	}
	if s == "" {
		return nil, nil
	}
	if !aimerMoney.MatchString(s) {
		return nil, d.Invalid("AimerHQ: confirm ambiguous amount")
	}
	return d.DecimalCents(strings.ReplaceAll(s, ",", ""))
}

func aimerDate(s, format string, epoch bool) (string, error) {
	if format == "dmy" || format == "mdy" {
		layout := "2/1/2006"
		if format == "mdy" {
			layout = "1/2/2006"
		}
		value, e := time.Parse(layout, s)
		if e == nil {
			return value.Format("2006-01-02"), nil
		}
	}
	return normalizedDate(s, format, epoch)
}

func aimerNonWorkRow(cells []string, m Mapping) bool {
	values := map[string]string{}
	for _, c := range m.Columns {
		if int(c.Column) < len(cells) {
			values[c.Field] = strings.TrimSpace(cells[c.Column])
		}
	}
	// Example and totals have no date, customer or work description. A partly
	// entered work row remains in the preview with errors instead of disappearing.
	if values["service_date"] != "" || values["customer"] != "" {
		return false
	}
	if strings.EqualFold(values["note"], "Example") {
		return true
	}
	return values["job_type"] == "" && values["detail"] == "" && values["duration"] == ""
}
