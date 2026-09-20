package domain

import (
	"regexp"
	"strings"
)

var aimerFuel = regexp.MustCompile(`(?i)\b(gas|fuel)\b`)

// ApplyAimerHQ only transforms a local snapshot. Company gross is never split
// or recalculated; the calculation applies exclusively when gross is absent.
func ApplyAimerHQ(d Data) (Data, error) {
	d.TeamSize = 1
	d.Participants = []string{"Jian"}
	name := strings.Join(strings.Fields(strings.ToLower(d.Customer)), " ")
	job := strings.Join(strings.Fields(strings.ToLower(d.JobType)), " ")
	for _, customer := range []string{"zhang's homestyle 张家巷", "thebasebar&grill 聚点", "chao's manukau", "mad dogs and englishmen", "tan's kitchen", "ace ktv & live house"} {
		if name == customer {
			d.TeamSize = 2
			d.Participants = []string{"Steven", "Jian"}
		}
	}
	if (d.ServiceDate == "2026-08-17" && name == "tingtea-albany 亭子茶" && job == "kiosk replacement") || (d.ServiceDate == "2026-08-18" && name == "milestone bar" && job == "pos collection") {
		d.TeamSize = 2
		d.Participants = []string{"Steven", "Jian"}
	}
	if d.ServiceDate == "2026-08-06" && name == "gogocow" && job == "pos installation" {
		d.TeamSize = 3
		d.Participants = []string{"Zhengyu", "Steven", "Jian"}
	}
	if d.ServiceDate == "2026-08-28" && name == "yes pancakes 也是馅饼" && job == "pos replacement" {
		d.TeamSize = 2
		d.Participants = []string{"Bin Lin", "Jian"}
	}
	return calculateAimerHQ(d)
}

func calculateAimerHQ(d Data) (Data, error) {
	if e := d.Validate(); e != nil {
		return d, e
	}
	job := strings.Join(strings.Fields(strings.ToLower(d.JobType)), " ")
	installation := job == "pos installation" || job == "kiosk installation"
	if d.Gross != nil {
		d.WageBasis = "company"
	} else {
		if d.DurationMinutes <= 0 {
			return d, Invalid("AimerHQ: confirm missing or zero job duration before estimating wages")
		}
		var project int64
		if installation {
			project = 16000
			if d.DurationMinutes > 300 {
				project += (d.DurationMinutes - 300) * 50
			}
		} else {
			// All remaining work is on-site, confirmed by the user on 2026-09-14.
			if job == "" {
				return d, Invalid("AimerHQ: job type is required before estimating wages")
			}
			minutes := max(int64(60), min(d.DurationMinutes, int64(120)))
			project = minutes * 50
		}
		gross := (project + d.TeamSize/2) / d.TeamSize
		d.Gross = &gross
		d.WageBasis = "rule_estimate"
	}
	if installation && d.Expense != nil && *d.Expense > 0 {
		note := strings.ToLower(d.Note)
		fuel := aimerFuel.MatchString(note)
		if fuel {
			if strings.Trim(aimerFuel.ReplaceAllString(note, ""), " ,;+&./\t\n") != "" {
				return d, Invalid("AimerHQ: confirm the fuel and other expense amounts separately")
			}
			zero := int64(0)
			d.Expense = &zero
		}
	}
	return d, d.Validate()
}
