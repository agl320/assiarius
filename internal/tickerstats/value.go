package tickerstats

import (
	"strconv"
	"strings"
)

type TokenKind int

// enum
const (
	TokenMissing TokenKind = iota
	TokenSeparator
	TokenText
	TokenNumber
)

type Token struct {
	Kind      TokenKind
	Raw       string
	Number    float64
	IsPercent bool
}

// Value is a parsed Finviz snapshot value.
//
// It keeps the original Raw string, but also exposes structured tokens so
// callers can easily pull out numbers/percents (including multi-part values
// like "4.95 -85.61%" or "14.48% 9.43%").
//
// Finviz uses "-" for missing values.
//
// Examples:
//   "-" -> Missing
//   "274.86M" -> Number(274860000)
//   "9.03%" -> Percent(9.03)
//   "4.95 -85.61%" -> Number(4.95), Percent(-85.61)
//   "Yes / Yes" -> Text("Yes"), Sep("/"), Text("Yes")
//   "Mar 31 AMC" -> Text tokens
//
// Note: numeric suffix multipliers are applied (K, M, B, T).
// Percent tokens store the numeric percent (e.g. 9.03 for "9.03%"), not a ratio.
//
// This type is intentionally generic; it does not try to infer semantics like
// "weekly vs monthly" for two-percent values.
//
type Value struct {
	Raw    string
	Tokens []Token
}

func ParseValue(raw string) Value {
	raw = strings.TrimSpace(strings.Join(strings.Fields(raw), " "))
	if raw == "" {
		return Value{Raw: raw, Tokens: []Token{{Kind: TokenMissing, Raw: ""}}}
	}
	if raw == "-" {
		return Value{Raw: raw, Tokens: []Token{{Kind: TokenMissing, Raw: raw}}}
	}

	fields := strings.Fields(raw)
	tokens := make([]Token, 0, len(fields))
	for _, f := range fields {
		if f == "-" {
			tokens = append(tokens, Token{Kind: TokenMissing, Raw: f})
			continue
		}
		if f == "/" {
			tokens = append(tokens, Token{Kind: TokenSeparator, Raw: f})
			continue
		}
		if tok, ok := parseNumberToken(f); ok {
			tokens = append(tokens, tok)
			continue
		}
		tokens = append(tokens, Token{Kind: TokenText, Raw: f})
	}

	return Value{Raw: raw, Tokens: tokens}
}

func (v Value) Missing() bool {
	if len(v.Tokens) == 0 {
		return true
	}
	for _, t := range v.Tokens {
		if t.Kind != TokenMissing && t.Kind != TokenSeparator {
			return false
		}
	}
	return true
}

func (v Value) Numbers() []float64 {
	out := make([]float64, 0, 2)
	for _, t := range v.Tokens {
		if t.Kind == TokenNumber && !t.IsPercent {
			out = append(out, t.Number)
		}
	}
	return out
}

func (v Value) Percents() []float64 {
	out := make([]float64, 0, 2)
	for _, t := range v.Tokens {
		if t.Kind == TokenNumber && t.IsPercent {
			out = append(out, t.Number)
		}
	}
	return out
}

func (v Value) FirstNumber() (float64, bool) {
	for _, t := range v.Tokens {
		if t.Kind == TokenNumber && !t.IsPercent {
			return t.Number, true
		}
	}
	return 0, false
}

func (v Value) FirstPercent() (float64, bool) {
	for _, t := range v.Tokens {
		if t.Kind == TokenNumber && t.IsPercent {
			return t.Number, true
		}
	}
	return 0, false
}

// NumberAndPercent returns the first non-percent number followed by the first
// percent token occurring after it.
func (v Value) NumberAndPercent() (number float64, percent float64, ok bool) {
	seenNumber := false
	for _, t := range v.Tokens {
		if t.Kind != TokenNumber {
			continue
		}
		if !t.IsPercent && !seenNumber {
			seenNumber = true
			number = t.Number
			continue
		}
		if t.IsPercent && seenNumber {
			return number, t.Number, true
		}
	}
	return 0, 0, false
}

func (v Value) TextTokens() []string {
	out := make([]string, 0, 4)
	for _, t := range v.Tokens {
		if t.Kind == TokenText {
			out = append(out, t.Raw)
		}
	}
	return out
}

func parseNumberToken(s string) (Token, bool) {
	orig := s
	if strings.HasSuffix(orig, ",") {
		return Token{}, false
	}
	isPercent := false
	if strings.HasSuffix(s, "%") {
		isPercent = true
		s = strings.TrimSuffix(s, "%")
	}

	mult := 1.0
	if len(s) > 0 {
		last := s[len(s)-1]
		switch last {
		case 'K', 'k':
			mult = 1_000
			s = s[:len(s)-1]
		case 'M', 'm':
			mult = 1_000_000
			s = s[:len(s)-1]
		case 'B', 'b':
			mult = 1_000_000_000
			s = s[:len(s)-1]
		case 'T', 't':
			mult = 1_000_000_000_000
			s = s[:len(s)-1]
		}
	}

	// strip commas (e.g. 13,494,076)
	s = strings.ReplaceAll(s, ",", "")
	if s == "" {
		return Token{}, false
	}

	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return Token{}, false
	}
	return Token{Kind: TokenNumber, Raw: orig, Number: v * mult, IsPercent: isPercent}, true
}
