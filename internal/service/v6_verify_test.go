package service

import "testing"

func TestV6ImportRemapsAccountRelation(t *testing.T) {
	oldID := uint(7)
	got := remapAccountID(&oldID, map[uint]uint{7: 19})
	if got == nil || *got != 19 {
		t.Fatalf("remapped account id = %v, want 19", got)
	}
}
