package server

import (
	"strings"
	"testing"
)

func TestRommetAxisCardsDoNotShareTheAbsoluteOverlayClass(t *testing.T) {
	body := getRommet(t)
	if !strings.Contains(body, `<section class="rom-axis-section">`) {
		t.Error("axis cards lost their in-flow section class")
	}
	if strings.Contains(body, `<section class="rom-axes">`) {
		t.Error("axis cards reuse the canvas overlay class and leave the document flow")
	}
	if !strings.Contains(body, `<div id="rommet-axes" class="rom-axes"`) {
		t.Error("the canvas lost its absolute axis overlay")
	}
}
