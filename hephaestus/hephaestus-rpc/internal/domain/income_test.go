package domain

import (
	"encoding/json"
	"testing"
)

func ptr(n int64) *int64 { return &n }
func TestMoneyAndUnknownValues(t *testing.T) {
	records := []Record{{Data: Data{Gross: ptr(3), Expense: ptr(0), DurationMinutes: 10}}, {Data: Data{Gross: ptr(3), Expense: ptr(0), DurationMinutes: 20}}, {Data: Data{Expense: ptr(25), DurationMinutes: 30}}, {Data: Data{Gross: ptr(100), DurationMinutes: 40}}}
	s, e := Total(records, 2000)
	if e != nil || s.Tax != 22 || s.Gross != 106 || s.Cash != 4 || s.Expense != 25 || s.Pending != 1 || s.UnknownExpense != 1 || s.Complete || s.Minutes != 100 || s.ConfirmedMinutes != 70 {
		t.Fatalf("unexpected summary: %+v %v", s, e)
	}
	if _, e = Total(nil, 10001); e == nil {
		t.Fatal("invalid rate accepted")
	}
	for _, raw := range []string{"-1", "1e3", "NaN", "1.234", "01.2", "1000000000.00"} {
		if _, e := DecimalCents(raw); e == nil {
			t.Fatalf("accepted %q", raw)
		}
	}
	v, e := DecimalCents("1021.00")
	if e != nil || *v != 102100 {
		t.Fatal(v, e)
	}
}
func TestOverridesAreIndependentOfSource(t *testing.T) {
	r := Record{SourceID: 1, Base: Data{ServiceDate: "2026-08-01", TeamSize: 1, Gross: ptr(100), Expense: ptr(0)}}
	if e := r.Refresh(); e != nil {
		t.Fatal(e)
	}
	if e := r.Edit([]Change{{Field: "gross_cents", Value: "200"}}, ""); e == nil {
		t.Fatal("missing reason accepted")
	}
	if e := r.Edit([]Change{{Field: "gross_cents", Value: "200"}}, "verified"); e != nil {
		t.Fatal(e)
	}
	r.Base.Gross = ptr(300)
	if e := r.Refresh(); e != nil || *r.Data.Gross != 200 || *r.Base.Gross != 300 {
		t.Fatal("source overwrote local correction")
	}
	if e := r.Edit([]Change{{Field: "gross_cents", Clear: true}}, "unknown"); e != nil || r.Data.Gross != nil || r.WageStatus != "pending" {
		t.Fatal(e)
	}
	b, _ := json.Marshal(r.Data)
	var decoded Data
	if e := json.Unmarshal(b, &decoded); e != nil || decoded.Gross != nil {
		t.Fatal("null money roundtrip failed")
	}
}
func TestBusinessDatesAndValidation(t *testing.T) {
	for _, v := range []string{"2026-02-29", "2026-2-01", "2026-09-31", "0000-01-01"} {
		if _, e := Date(v); e == nil {
			t.Fatal(v)
		}
	}
	if _, e := Date("2028-02-29"); e != nil {
		t.Fatal(e)
	}
	d := Data{ServiceDate: "2026-09-09", TeamSize: 1}
	if _, e := Apply(d, []Change{{Field: "duration_minutes", Value: "1441"}}); e == nil {
		t.Fatal("duration over daily bound")
	}
	if _, e := Apply(d, []Change{{Field: "gross_cents", Value: "0"}, {Field: "gross_cents", Clear: true}}); e == nil {
		t.Fatal("duplicate patch")
	}
}
