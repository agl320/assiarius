package screener

import (
	"strings"
	"testing"
	"time"
)

func TestValidateFinvizScreenerURL(t *testing.T) {
	t.Parallel()

	valid := "https://finviz.com/screener.ashx?v=111&f=news_date_prevminutes5,sh_curvol_o500,sh_price_u5,sh_relvol_o1&ft=6"
	validNoF := "https://finviz.com/screener.ashx?v=111&ft=6"
	validRegularFilters := "https://finviz.com/screener.ashx?v=111&f=sh_avgvol_u100"

	tests := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{"empty", "", true},
		{"not_url", "not a url", true},
		{"missing_scheme", "finviz.com/screener.ashx?v=111&f=news_date_prevminutes5", true},
		{"wrong_host", "https://example.com/screener.ashx?v=111&f=news_date_prevminutes5", true},
		{"wrong_path", "https://finviz.com/quote.ashx?t=TSLA", true},
		{"missing_f", "https://finviz.com/screener.ashx?v=111", false},
		{"valid_no_f", validNoF, false},
		{"valid_regular_filters", validRegularFilters, false},
		{"no_news_window", "https://finviz.com/screener.ashx?v=111&f=sh_curvol_o500,sh_price_u5&ft=6", false},
		{"valid", valid, false},
		{"valid_encoded", "https://finviz.com/screener.ashx?v=111&f=news_date_prevminutes5%2Csh_curvol_o500%2Csh_price_u5&ft=6", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := ValidateFinvizScreenerURL(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err=%v wantErr=%v", err, tt.wantErr)
			}
		})
	}
}

func TestBuildNewsScreenerURL_ReplacesWindowFilter(t *testing.T) {
	t.Parallel()

	base := "https://finviz.com/screener.ashx?v=111&f=news_date_prevhours24,sh_curvol_o500,sh_price_u5,sh_relvol_o1&ft=6"

	out, err := BuildNewsScreenerURL(base, 5*time.Minute)
	if err != nil {
		t.Fatalf("BuildNewsScreenerURL returned error: %v", err)
	}
	if !strings.Contains(out, "news_date_prevminutes5") {
		t.Fatalf("expected minutes filter in output: %s", out)
	}
	if strings.Contains(out, "news_date_prevhours24") {
		t.Fatalf("expected old hours filter to be removed: %s", out)
	}
	// Ensure other filters preserved.
	for _, f := range []string{"sh_curvol_o500", "sh_price_u5", "sh_relvol_o1"} {
		if !strings.Contains(out, f) {
			t.Fatalf("expected filter %q preserved in output: %s", f, out)
		}
	}
}

func TestBuildNewsScreenerURL_AddsWindowFilterWhenMissing(t *testing.T) {
	t.Parallel()

	base := "https://finviz.com/screener.ashx?v=111&f=sh_curvol_o500,sh_price_u5,sh_relvol_o1&ft=6"
	out, err := BuildNewsScreenerURL(base, 5*time.Minute)
	if err != nil {
		t.Fatalf("BuildNewsScreenerURL returned error: %v", err)
	}
	if !strings.Contains(out, "news_date_prevminutes5") {
		t.Fatalf("expected minutes filter in output: %s", out)
	}
}

func TestBuildNewsScreenerURL_AddsWindowFilterWhenFIsMissing(t *testing.T) {
	t.Parallel()

	base := "https://finviz.com/screener.ashx?v=111&ft=6"
	out, err := BuildNewsScreenerURL(base, 30*time.Minute)
	if err != nil {
		t.Fatalf("BuildNewsScreenerURL returned error: %v", err)
	}
	if !strings.Contains(out, "news_date_prevminutes30") {
		t.Fatalf("expected minutes filter in output: %s", out)
	}
}

func TestBuildNewsScreenerURL_HoursWindow(t *testing.T) {
	t.Parallel()

	base := "https://finviz.com/screener.ashx?v=111&f=news_date_prevminutes5,sh_curvol_o500&ft=6"
	out, err := BuildNewsScreenerURL(base, 24*time.Hour)
	if err != nil {
		t.Fatalf("BuildNewsScreenerURL returned error: %v", err)
	}
	if !strings.Contains(out, "news_date_prevhours24") {
		t.Fatalf("expected hours filter in output: %s", out)
	}
	if strings.Contains(out, "news_date_prevminutes5") {
		t.Fatalf("expected old minutes filter to be removed: %s", out)
	}
}

func TestBuildNewsScreenerURL_InvalidWindow(t *testing.T) {
	t.Parallel()

	base := "https://finviz.com/screener.ashx?v=111&f=news_date_prevminutes5,sh_curvol_o500&ft=6"

	_, err := BuildNewsScreenerURL(base, 90*time.Second)
	if err == nil {
		t.Fatalf("expected error for non-minute/hour window")
	}

	_, err = BuildNewsScreenerURL(base, 0)
	if err == nil {
		t.Fatalf("expected error for zero window")
	}
}
