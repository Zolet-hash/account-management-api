package postgres

import "testing"

func TestCentsUsesExactDecimalParsing(t *testing.T) {
	for input, want := range map[string]int64{"0.01": 1, "-0.50": -50, "42": 4200, "10.2": 1020, "9999999999999999.99": 999999999999999999} {
		got, err := tocent(input)
		if err != nil || got != want {
			t.Errorf("cents(%q) = %d, %v; want %d", input, got, err, want)
		}
	}
	for _, input := range []string{"1.001", "x", "92233720368547758.08"} {
		if _, err := tocent(input); err == nil {
			t.Errorf("cents(%q) unexpectedly succeeded", input)
		}
	}
}

func TestMoneyPreservesNegativeMinorAmounts(t *testing.T) {
	for cents, want := range map[int64]string{-1: "-0.01", -50: "-0.50", -150: "-1.50", 50: "0.50"} {
		if got := money(cents); got != want {
			t.Errorf("money(%d) = %q; want %q", cents, got, want)
		}
	}
}