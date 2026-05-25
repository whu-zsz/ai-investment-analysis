package model

import "testing"

func TestNormalizeIndustryLabel_RejectsNumericValue(t *testing.T) {
	if got := NormalizeIndustryLabel("-2.21"); got != "" {
		t.Fatalf("NormalizeIndustryLabel() = %q, want empty", got)
	}
}

func TestNormalizeConceptList_RejectsNumericValue(t *testing.T) {
	got := NormalizeConceptList([]string{"52.2244889889", "白酒"})
	if len(got) != 1 || got[0] != "白酒" {
		t.Fatalf("NormalizeConceptList() = %#v, want [\"白酒\"]", got)
	}
}

func TestNormalizeConceptList_RejectsPlaceholderValue(t *testing.T) {
	got := NormalizeConceptList([]string{"-", "—", "沪股通"})
	if len(got) != 1 || got[0] != "沪股通" {
		t.Fatalf("NormalizeConceptList() = %#v, want [\"沪股通\"]", got)
	}
}
