package marketdata

import (
	"regexp"
	"testing"
)

func TestDecodeSinaEscapedText(t *testing.T) {
	got := decodeSinaEscapedText(`\u7535\u5b50\u5668\u4ef6`)
	if got != "电子器件" {
		t.Fatalf("decodeSinaEscapedText() = %q, want %q", got, "电子器件")
	}
}

func TestAppendMatchedNodes_DecodesNodeNames(t *testing.T) {
	text := `["\u7535\u5b50\u5668\u4ef6","","new_dzqj","cn"]`
	var nodes []MarketNode
	appendMatchedNodes(text, mustCompile(`\["([^"]+)","","(new_[^"]+)"(?:,"cn")?\]`), "industry", map[string]struct{}{}, &nodes)
	if len(nodes) != 1 {
		t.Fatalf("nodes len = %d, want 1", len(nodes))
	}
	if nodes[0].Name != "电子器件" {
		t.Fatalf("nodes[0].Name = %q, want %q", nodes[0].Name, "电子器件")
	}
}

func mustCompile(expr string) *regexp.Regexp {
	return regexp.MustCompile(expr)
}
