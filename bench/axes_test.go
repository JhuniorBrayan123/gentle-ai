package main

import "testing"

// The fork retains the OpenCode SDD transport axis as a supported benchmark.
func TestRetainedSDDTaskResultAxisIsRegistered(t *testing.T) {
	found := false
	for _, axis := range Axes() {
		if axis.Name == sddTaskResultAxis {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("retained SDD task-result transport axis is not registered")
	}
	selected, err := selectAxes(sddTaskResultAxis)
	if err != nil {
		t.Fatalf("select retained SDD task-result axis: %v", err)
	}
	if len(selected) != 1 || selected[0].Name != sddTaskResultAxis {
		t.Fatalf("selected axes = %+v, want %q", selected, sddTaskResultAxis)
	}
}
