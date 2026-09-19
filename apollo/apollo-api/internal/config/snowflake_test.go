package config

import "testing"

func TestResolveSnowflakeNodeID(t *testing.T) {
	if got, err := ResolveSnowflakeNodeID(321); err != nil || got != 321 {
		t.Fatalf("explicit node: %d %v", got, err)
	}
	for _, value := range []int64{-1, 1024} {
		if _, err := ResolveSnowflakeNodeID(value); err == nil {
			t.Fatalf("invalid node accepted: %d", value)
		}
	}
	for podIP, want := range map[string]int64{
		"10.42.7.21": 789,
		"10.42.7.22": 790,
		"10.42.7.23": 791,
	} {
		t.Setenv("POD_IP", podIP)
		if got, err := ResolveSnowflakeNodeID(0); err != nil || got != want {
			t.Fatalf("pod %s node: got %d, want %d, err %v", podIP, got, want, err)
		}
	}
}

func TestResolveSnowflakeNodeIDRejectsInvalidPodIP(t *testing.T) {
	for _, value := range []string{"invalid", "2001:db8::1"} {
		t.Setenv("POD_IP", value)
		if _, err := ResolveSnowflakeNodeID(0); err == nil {
			t.Fatalf("invalid POD_IP accepted: %q", value)
		}
	}
}
