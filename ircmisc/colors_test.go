package ircmisc

import (
	"testing"
)

func TestConsumeText(t *testing.T) {
	s := "abc\x02123"
	consumed, remaining := consumeText(s)
	if consumed != "abc" {
		t.Fatalf("expected abc, got %v", consumed)
	}

	if remaining != "\x02123" {
		t.Fatalf("r1: expected 123, got %v (%d)", remaining, len(remaining))
	}

	consumed, remaining = consumeText(remaining[1:])
	if consumed != "123" {
		t.Fatalf("r2: expected 123 got, %v", consumed)
	}

	if remaining != "" {
		t.Fatalf("expected no more remaining, got %v", remaining)
	}

	consumed, remaining = consumeText("")
	if consumed != "" || remaining != "" {
		t.Fatalf("expected empties")
	}
}

func TestFormatting(t *testing.T) {
	var f Formatting
	f.SetBackgroundColor(13)
	f.SetForegroundColor(3)

	if f.ForegroundColor() != 3 {
		t.Fatalf("invalid foreground color")
	}

	if f.BackgroundColor() != 13 {
		t.Fatalf("invalid foreground color")
	}
}

func TestTokenizeFormattingNothing(t *testing.T) {
	var (
		r []FormattedToken
	)

	r = TokenizeByFormatting("no codes here")
	if len(r) != 1 {
		t.Fatalf("expected 1 token, got %v", r)
	}

	if r[0].Text != "no codes here" {
		t.Fatalf("expected 'no codes here' got %v", r[0])
	}
}

func TestTokenizeFormattingToggles(t *testing.T) {
	var (
		r []FormattedToken
	)

	r = TokenizeByFormatting("This text has \x02bolded\x02 text, \x1funderlined text\x1f, and a \x02\x1f\x1ecombination thereof\x1e\x1f\x02.")
	if len(r) != 7 {
		t.Fatalf("unexpected number of tokens")
	}

	expectedTokens := []string{"This text has ", "bolded", " text, ", "underlined text", ", and a ", "combination thereof", "."}
	for i := 0; i < len(r); i++ {
		if expectedTokens[i] != r[i].Text {
			t.Errorf("invalid token %d, expected %v got %v", i, expectedTokens[i], r[i].Text)
		}
	}

	bolded := []bool{false, true, false, false, false, true, false}
	for i := 0; i < len(r); i++ {
		if r[i].Format.IsBold() != bolded[i] {
			t.Errorf("expected token %d to be bold=%v, instead got %v", i, bolded[i], r[i].Format.IsBold())
		}
	}
}

// examples taken from https://modern.ircdocs.horse/formatting.html
func TestTokenizeColors1(t *testing.T) {
	var (
		text string
		r    []FormattedToken
	)
	text = "I love \x033 IRC! \x03It is the \x037best protocol ever!"
	r = TokenizeByFormatting(text)
	if len(r) != 4 {
		t.Fatalf("expected 4 tokens, got %d", len(text))
	}

	if r[1].Format.ForegroundColor() != 3 {
		t.Fatalf("token 1 should have foreground color 3, got %d", r[1].Format.ForegroundColor())
	}

	if r[2].Format.ForegroundColor() != 0 {
		t.Fatalf("token 1 should have foreground color 0, got %d", r[2].Format.ForegroundColor())
	}

	if r[3].Format.ForegroundColor() != 7 {
		t.Fatalf("token 3 should have foreground color 7, got %d", r[3].Format.ForegroundColor())
	}

}

func TestTokenizeColors2(t *testing.T) {
	var (
		text string
		r    []FormattedToken
	)
	text = "Rules: Don't spam 5\x0313,8,6\x03,7,8, and especially not \x029\x02\x1D!"
	r = TokenizeByFormatting(text)
	if len(r) != 5 {
		t.Fatalf("expected 5 tokens, got %d", len(r))
	}

	expectedTokens := []string{"Rules: Don't spam 5", ",6", ",7,8, and especially not ", "9", "!"}
	for i := 0; i < len(expectedTokens); i++ {
		if r[i].Text != expectedTokens[i] {
			t.Errorf("error on token %d. expected %v got %v", i, expectedTokens[i], r[i].Text)
		}
	}

	expectedFg := []int{0, 13, 0, 0, 0}
	expectedBg := []int{0, 8, 0, 0, 0}
	for i := 0; i < len(r); i++ {
		efg := expectedFg[i]
		fg := r[i].Format.ForegroundColor()
		if efg != fg {
			t.Errorf("token %d expected fg %d got %d", i, efg, fg)
		}

		ebg := expectedBg[i]
		bg := r[i].Format.BackgroundColor()
		if ebg != bg {
			t.Errorf("token %d expected bg %d got %d", i, ebg, bg)
		}
	}
}
