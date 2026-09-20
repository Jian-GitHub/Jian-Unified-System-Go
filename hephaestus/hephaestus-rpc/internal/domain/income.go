package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type Error struct{ Code, Message string }

func (e *Error) Error() string        { return e.Message }
func Fail(code, message string) error { return &Error{code, message} }
func Invalid(message string) error    { return Fail("INVALID_ARGUMENT", message) }
func Conflict() error                 { return Fail("VERSION_CONFLICT", "record or preview has changed") }

var digits = regexp.MustCompile(`^(0|[1-9][0-9]*)$`)

const MaxMoney int64 = 99999999999

func ID(s string) (int64, error) {
	n, e := strconv.ParseInt(s, 10, 64)
	if e != nil || n <= 0 || strconv.FormatInt(n, 10) != s {
		return 0, Invalid("invalid identifier")
	}
	return n, nil
}
func Date(s string) (time.Time, error) {
	t, e := time.Parse("2006-01-02", s)
	if e != nil || t.Format("2006-01-02") != s || t.Year() < 1901 || t.Year() > 9999 {
		return t, Invalid("invalid business date")
	}
	return t, nil
}
func Cents(s string) (*int64, error) {
	if !digits.MatchString(s) {
		return nil, Invalid("money must be a nonnegative integer cent string")
	}
	n, e := strconv.ParseInt(s, 10, 64)
	if e != nil || n > MaxMoney {
		return nil, Invalid("money exceeds limit")
	}
	return &n, nil
}
func Hash(v any) string {
	b, _ := json.Marshal(v)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

type Data struct {
	WageBasis       string   `json:"wage_basis"`
	Participants    []string `json:"participants"`
	PaymentStatus   string   `json:"payment_status"`
	ServiceDate     string   `json:"service_date"`
	Customer        string   `json:"customer"`
	JobType         string   `json:"job_type"`
	Detail          string   `json:"detail"`
	Note            string   `json:"note"`
	PricingCategory string   `json:"pricing_category"`
	DurationMinutes int64    `json:"duration_minutes"`
	TeamSize        int64    `json:"team_size"`
	Gross           *int64   `json:"gross_cents,string"`
	Expense         *int64   `json:"expense_cents,string"`
}

func (d Data) Validate() error {
	if d.WageBasis != "" && d.WageBasis != "company" && d.WageBasis != "rule_estimate" {
		return Invalid("invalid wage basis")
	}
	if len(d.PaymentStatus) > 200 || strings.ContainsRune(d.PaymentStatus, '\x00') {
		return Invalid("invalid payment status")
	}
	if _, e := Date(d.ServiceDate); e != nil {
		return e
	}
	for _, s := range []string{d.Customer, d.JobType, d.PricingCategory, d.Detail, d.Note} {
		if !utf8.ValidString(s) || strings.ContainsAny(s, "\x00") {
			return Invalid("invalid text")
		}
	}
	if utf8.RuneCountInString(d.Customer) > 200 || utf8.RuneCountInString(d.JobType) > 100 || utf8.RuneCountInString(d.PricingCategory) > 100 || utf8.RuneCountInString(d.Detail) > 2000 || utf8.RuneCountInString(d.Note) > 2000 {
		return Invalid("text exceeds limit")
	}
	if d.DurationMinutes < 0 || d.DurationMinutes > 1440 || d.TeamSize < 1 || d.TeamSize > 1000 {
		return Invalid("duration must be 0..1440 minutes and team size 1..1000")
	}
	for _, p := range []*int64{d.Gross, d.Expense} {
		if p != nil && (*p < 0 || *p > MaxMoney) {
			return Invalid("money exceeds limit")
		}
	}
	return nil
}

type Change struct {
	Field string `json:"field"`
	Value string `json:"value"`
	Clear bool   `json:"clear"`
}

func Apply(base Data, changes []Change) (Data, error) {
	d := base
	seen := map[string]bool{}
	for _, c := range changes {
		if seen[c.Field] {
			return d, Invalid("duplicate changed field")
		}
		seen[c.Field] = true
		if c.Clear && c.Value != "" {
			return d, Invalid("clear and value are mutually exclusive")
		}
		if c.Clear && c.Field != "gross_cents" && c.Field != "expense_cents" {
			return d, Invalid("only unknown monetary values can be cleared")
		}
		switch c.Field {
		case "service_date":
			d.ServiceDate = c.Value
		case "customer":
			d.Customer = c.Value
		case "job_type":
			d.JobType = c.Value
		case "detail":
			d.Detail = c.Value
		case "note":
			d.Note = c.Value
		case "pricing_category":
			d.PricingCategory = c.Value
		case "duration_minutes", "team_size":
			n, e := strconv.ParseInt(c.Value, 10, 64)
			if e != nil {
				return d, Invalid("invalid integer")
			}
			if c.Field == "team_size" {
				if n != d.TeamSize {
					d.Participants = nil
				}
				d.TeamSize = n
			} else {
				d.DurationMinutes = n
			}
		case "gross_cents", "expense_cents":
			var p *int64
			var e error
			if !c.Clear {
				p, e = Cents(c.Value)
				if e != nil {
					return d, e
				}
			}
			if c.Field == "gross_cents" {
				d.Gross = p
			} else {
				d.Expense = p
			}
		default:
			return d, Invalid("unknown changed field")
		}
	}
	return d, d.Validate()
}

type Record struct {
	ID         int64    `json:"id,string"`
	Version    int64    `json:"version,string"`
	Data       Data     `json:"data"`
	Base       Data     `json:"base"`
	Overrides  []Change `json:"overrides"`
	SourceID   int64    `json:"source_id,string"`
	SourceKey  string   `json:"source_key"`
	Archived   bool     `json:"archived"`
	WageStatus string   `json:"wage_status"`
}

func (r *Record) Refresh() error {
	d, e := Apply(r.Base, r.Overrides)
	if e != nil {
		return e
	}
	r.Data = d
	if r.Base.WageBasis == "company" {
		r.Data.Gross = r.Base.Gross
		r.Data.WageBasis = "company"
	}
	if r.Base.WageBasis == "rule_estimate" {
		r.Data.Gross = nil
		r.Data, e = calculateAimerHQ(r.Data)
		if e != nil {
			return e
		}
	}
	d = r.Data
	r.WageStatus = "pending"
	if d.Gross != nil {
		r.WageStatus = "confirmed"
		if d.WageBasis == "rule_estimate" {
			r.WageStatus = "estimated"
		}
	}
	if r.Overrides == nil {
		r.Overrides = []Change{}
	}
	return nil
}
func (r *Record) Edit(changes []Change, reason string) error {
	if r.Archived {
		return Fail("CONFLICT", "restore the archived record before editing")
	}
	if len(changes) == 0 || len(changes) > 10 {
		return Invalid("provide 1..10 changes")
	}
	if _, e := Apply(r.Data, changes); e != nil {
		return e
	}
	if r.SourceID == 0 {
		d, e := Apply(r.Base, changes)
		if e != nil {
			return e
		}
		r.Base = d
	} else {
		if strings.TrimSpace(reason) == "" {
			return Invalid("a reason is required for imported record edits")
		}
		for _, c := range changes {
			if c.Field == "gross_cents" && r.Base.WageBasis != "" {
				return Invalid("AimerHQ wages come from the company or confirmed estimation rules; reimport company wages instead")
			}
			found := false
			for i, o := range r.Overrides {
				if o.Field == c.Field {
					r.Overrides[i] = c
					found = true
					break
				}
			}
			if !found {
				r.Overrides = append(r.Overrides, c)
			}
		}
	}
	return r.Refresh()
}

type Summary struct {
	Gross              int64  `json:"gross_cents,string"`
	Tax                int64  `json:"tax_estimate_cents,string"`
	Net                int64  `json:"net_estimate_cents,string"`
	Expense            int64  `json:"known_expense_cents,string"`
	Cash               int64  `json:"known_cash_estimate_cents,string"`
	Count              int64  `json:"record_count"`
	Pending            int64  `json:"pending_wage_count"`
	UnknownExpense     int64  `json:"unknown_expense_count"`
	Minutes            int64  `json:"duration_minutes"`
	ConfirmedMinutes   int64  `json:"confirmed_duration_minutes"`
	Complete           bool   `json:"is_complete"`
	CalculationVersion string `json:"calculation_version"`
}

func Total(records []Record, rate int64) (Summary, error) {
	s := Summary{Complete: true, CalculationVersion: "income-v1"}
	if rate < 0 || rate > 10000 {
		return s, Invalid("rate_bps must be 0..10000")
	}
	if len(records) > 100000 {
		return s, Invalid("narrow the reporting range to 100000 records")
	}
	for _, r := range records {
		d := r.Data
		s.Count++
		if d.WageBasis == "rule_estimate" {
			s.Pending++
			s.Complete = false
		}
		s.Minutes += d.DurationMinutes
		if d.Expense == nil {
			s.UnknownExpense++
			s.Complete = false
		} else {
			s.Expense += *d.Expense
		}
		if d.Gross == nil {
			s.Pending++
			s.Complete = false
			continue
		}
		wageRate := rate
		if d.WageBasis != "" {
			wageRate = 2000
		}
		tax := (*d.Gross*wageRate + 5000) / 10000
		s.Gross += *d.Gross
		s.Tax += tax
		s.Net += *d.Gross - tax
		if d.WageBasis != "rule_estimate" {
			s.ConfirmedMinutes += d.DurationMinutes
		}
		if d.Expense != nil {
			s.Cash += *d.Gross - tax + *d.Expense
		}
	}
	return s, nil
}

type Filter struct {
	From, To, Type, Team, Search, Sort, Archived, Cursor, GroupBy string
	Limit, Rate                                                   int64
}

func (f Filter) Validate() error {
	if f.From != "" {
		if _, e := Date(f.From); e != nil {
			return e
		}
	}
	if f.To != "" {
		if _, e := Date(f.To); e != nil {
			return e
		}
	}
	if f.From != "" && f.To != "" && f.From > f.To {
		return Invalid("from must not be later than to")
	}
	if f.Rate < 0 || f.Rate > 10000 {
		return Invalid("rate_bps must be 0..10000")
	}
	if f.Limit < 1 || f.Limit > 200 {
		return Invalid("limit must be 1..200")
	}
	if !oneOf(f.Team, "", "all", "solo", "team") || !oneOf(f.Sort, "", "date-desc", "date-asc", "gross-desc", "cash-desc") || !oneOf(f.Archived, "", "active", "archived", "all") || !oneOf(f.GroupBy, "", "day", "week", "month") {
		return Invalid("unsupported query option")
	}
	if len(f.Search) > 200 || len(f.Type) > 100 {
		return Invalid("query too long")
	}
	return nil
}
func oneOf(s string, vs ...string) bool {
	for _, v := range vs {
		if s == v {
			return true
		}
	}
	return false
}
func Reason(s string) error {
	if utf8.RuneCountInString(s) > 1000 {
		return Invalid("reason exceeds limit")
	}
	return nil
}
func Version(actual, expected int64) error {
	if expected < 1 {
		return Fail("PRECONDITION_REQUIRED", "If-Match is required")
	}
	if actual != expected {
		return Conflict()
	}
	return nil
}
func DecimalCents(s string) (*int64, error) {
	if s == "" {
		return nil, nil
	}
	if !regexp.MustCompile(`^(0|[1-9][0-9]*)(\.[0-9]{1,2})?$`).MatchString(s) {
		return nil, Invalid("expected decimal amount with at most two places")
	}
	p := strings.Split(s, ".")
	tail := "00"
	if len(p) == 2 {
		tail = (p[1] + "0")[:2]
	}
	n, e := strconv.ParseInt(p[0], 10, 64)
	if e != nil || n > MaxMoney/100 {
		return nil, Invalid("money exceeds limit")
	}
	c, _ := strconv.ParseInt(tail, 10, 64)
	v := n*100 + c
	if v > MaxMoney {
		return nil, Invalid("money exceeds limit")
	}
	return &v, nil
}
func FormatMoney(p *int64) string {
	if p == nil {
		return ""
	}
	return fmt.Sprintf("%d.%02d", *p/100, *p%100)
}
func (e *Error) BusinessCode() string { return e.Code }
