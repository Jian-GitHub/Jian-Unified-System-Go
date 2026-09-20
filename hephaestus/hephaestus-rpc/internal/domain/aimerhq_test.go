package domain

import "testing"

func TestAimerWagePrecedenceAndEstimates(t *testing.T) {
	for _, tc := range []struct {
		job     string
		minutes int64
		gross   int64
	}{
		{"POS Installation", 120, 16000}, {"Kiosk Installation", 360, 19000},
		{"Printer Repair", 15, 3000}, {"Printer Installation", 180, 6000},
		{"CashDrawer Key Delivery", 60, 3000}, {"Site inspection", 180, 6000},
	} {
		d, e := ApplyAimerHQ(Data{ServiceDate: "2026-09-01", JobType: tc.job, DurationMinutes: tc.minutes, TeamSize: 1})
		if e != nil || *d.Gross != tc.gross || d.WageBasis != "rule_estimate" {
			t.Fatalf("%+v %v", d, e)
		}
	}
	company := int64(9000)
	cost := int64(1700)
	d, e := ApplyAimerHQ(Data{ServiceDate: "2026-08-05", Customer: "Chao's Manukau", JobType: "Printer Installation", DurationMinutes: 180, Gross: &company, Expense: &cost, TeamSize: 1})
	if e != nil || *d.Gross != 9000 || d.TeamSize != 2 || *d.Expense != 1700 || d.WageBasis != "company" {
		t.Fatalf("company altered: %+v %v", d, e)
	}
	r := Record{Base: d}
	if e = r.Refresh(); e != nil {
		t.Fatal(e)
	}
	s, e := Total([]Record{r}, 2000)
	if e != nil || s.Gross != 9000 || s.Tax != 1800 || s.Cash != 8900 {
		t.Fatalf("tax/cost: %+v %v", s, e)
	}
}

func TestAimerParticipantsAndAmbiguity(t *testing.T) {
	d := Data{ServiceDate: "2026-08-06", Customer: "GoGoCow", JobType: "POS Installation", DurationMinutes: 120, TeamSize: 1}
	v, e := ApplyAimerHQ(d)
	if e != nil || *v.Gross != 5333 || v.TeamSize != 3 {
		t.Fatalf("three person split: %+v %v", v, e)
	}
	d.ServiceDate = "2026-08-08"
	d.JobType = "Printer Installation"
	v, e = ApplyAimerHQ(d)
	if e != nil || v.TeamSize != 1 || *v.Gross != 6000 {
		t.Fatal(v, e)
	}
	d.Customer = "Yes Pancakes  也是馅饼"
	d.ServiceDate = "2026-08-28"
	d.JobType = "POS Replacement"
	v, e = ApplyAimerHQ(d)
	if e != nil || *v.Gross != 3000 || v.Participants[0] != "Bin Lin" {
		t.Fatal(v, e)
	}
	d.JobType = ""
	if _, e = ApplyAimerHQ(d); e == nil {
		t.Fatal("unknown job silently estimated")
	}
	d.JobType = "POS Installation"
	cost := int64(1500)
	d.Expense = &cost
	d.Note = "Gas, Parking"
	if _, e = ApplyAimerHQ(d); e == nil {
		t.Fatal("mixed cost silently split")
	}
	d.Note = "Fuel"
	v, e = ApplyAimerHQ(d)
	if e != nil || *v.Expense != 0 {
		t.Fatal(v, e)
	}
}
