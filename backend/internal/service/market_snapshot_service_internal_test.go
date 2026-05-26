package service

import "testing"

func TestDecodeBoardDisplayName(t *testing.T) {
	got := decodeBoardDisplayName(`\u7535\u5b50\u5668\u4ef6`)
	if got != "电子器件" {
		t.Fatalf("decodeBoardDisplayName() = %q, want %q", got, "电子器件")
	}
}

func TestBoardSearchTerms_ExtractsMeaningfulChineseTerms(t *testing.T) {
	terms := boardSearchTerms("我想买白酒相关股票")
	want := map[string]struct{}{
		"白酒": {},
	}
	for term := range want {
		found := false
		for _, item := range terms {
			if item == term {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("boardSearchTerms() missing %q in %#v", term, terms)
		}
	}
}
