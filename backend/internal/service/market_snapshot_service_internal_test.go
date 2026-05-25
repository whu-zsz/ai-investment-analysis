package service

import "testing"

func TestDecodeBoardDisplayName(t *testing.T) {
	got := decodeBoardDisplayName(`\u7535\u5b50\u5668\u4ef6`)
	if got != "电子器件" {
		t.Fatalf("decodeBoardDisplayName() = %q, want %q", got, "电子器件")
	}
}
