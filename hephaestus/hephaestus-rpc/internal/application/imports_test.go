package application

import "testing"

func TestDatesAndExactDuration(t *testing.T) {
	for _, tc := range []struct {
		in, format, out string
		epoch           bool
	}{{"61", "excel", "1900-03-01", false}, {"0", "excel", "1904-01-01", true}, {"28/08/2026", "dmy", "2026-08-28", false}, {"08/28/2026", "mdy", "2026-08-28", false}} {
		got, e := normalizedDate(tc.in, tc.format, tc.epoch)
		if tc.out == "1900-03-01" {
			if e == nil {
				t.Fatal("pre-1901 date outside supported domain accepted")
			}
			continue
		}
		if e != nil || got != tc.out {
			t.Fatal(tc, got, e)
		}
	}
	if _, e := normalizedDate("60", "excel", false); e == nil {
		t.Fatal("Excel fictitious date accepted")
	}
	for _, v := range []string{"1.001", "1e2", "-1", "1/2"} {
		if _, e := duration(v, "hours"); e == nil {
			t.Fatal(v)
		}
	}
	if v, e := duration("1.25", "hours"); e != nil || v != 75 {
		t.Fatal(v, e)
	}
	if v, e := duration("01:30", "hh:mm"); e != nil || v != 90 {
		t.Fatal(v, e)
	}
}
