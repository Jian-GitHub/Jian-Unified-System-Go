package identity

import (
	"testing"
	"time"
)

func TestProjectScopeCombinations(t *testing.T) {
	now := time.Now().UTC()
	for mask := int64(1); mask <= 7; mask++ {
		var scopes []int64
		for _, flag := range []int64{ScopeJQuantum, ScopeArgus, ScopeHephaestus} {
			if mask&flag != 0 {
				scopes = append(scopes, flag, flag)
			}
		}
		g, err := NewGrant(1, 2, "projects", scopes, now, time.Hour)
		if err != nil || g.Scope() != mask {
			t.Fatalf("mask %d: %v", mask, err)
		}
		restored, err := RestoreGrant(g.ID(), g.OwnerID(), g.Scope(), g.Name(), "signed", g.CreatedAt(), g.ExpiresAt(), true, false)
		if err != nil || restored.Scope() != mask || !restored.Active(2, now) {
			t.Fatalf("restore mask %d: %v", mask, err)
		}
	}
	for _, mask := range []int64{0, -1, 8, 9, 15} {
		if _, err := RestoreGrant(1, 2, mask, "projects", "signed", now, now.Add(time.Hour), true, false); err == nil {
			t.Fatalf("invalid mask %d accepted", mask)
		}
	}
}
