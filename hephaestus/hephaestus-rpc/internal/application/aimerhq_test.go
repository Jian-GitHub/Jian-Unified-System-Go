package application

import "testing"

func TestAimerMappingRequiresExplicitWageColumn(t *testing.T) {
	m := Mapping{RuleProfile: "aimerhq-v1", HeaderRow: 1, DateFormat: "excel", DurationFormat: "hours", DefaultTeamSize: 1, Columns: []Column{{"service_date", 0}, {"customer", 1}, {"job_type", 2}, {"detail", 3}, {"duration", 4}, {"note", 5}, {"expense", 6}, {"gross", 7}}}
	if validateMapping(m) == nil {
		t.Fatal("unconfirmed wage mapping accepted")
	}
	m.WageColumnConfirmed = true
	if e := validateMapping(m); e != nil {
		t.Fatal(e)
	}
	row := []string{"46234.0", "Zhang's Homestyle 张家巷", "Pos Installation", "details", "2.0", "", "", ""}
	v, k, e := normalize(row, m, false)
	if e != nil || v.ServiceDate != "2026-07-31" || *v.Gross != 8000 || v.WageBasis != "rule_estimate" {
		t.Fatal(v, e)
	}
	if v.Expense == nil || *v.Expense != 0 {
		t.Fatal("confirmed blank cost must be zero", v)
	}
	// The user's workbook puts wages in I, despite its Payment Status header.
	m.Columns[len(m.Columns)-1].Column = 8
	row[7] = "25.50"
	row = append(row, "123.45")
	v, k2, e := normalize(row, m, false)
	if e != nil || *v.Gross != 12345 || v.WageBasis != "company" || k != k2 {
		t.Fatal("company update must retain identity", v, e)
	}
	if !aimerNonWorkRow([]string{"", "", "Kiosk Installnation", "example", "", "Example"}, m) {
		t.Fatal("example included")
	}
	if !aimerNonWorkRow([]string{"", "", "", "", "", "", "", "254.66", "1021"}, m) {
		t.Fatal("total included")
	}
	if aimerNonWorkRow([]string{"", "Customer"}, m) {
		t.Fatal("incomplete work silently skipped")
	}
}
