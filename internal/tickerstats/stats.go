package tickerstats

import (
	"strings"
	"time"
)

type Stats struct {
	Ticker      string
	RetrievedAt time.Time

	byLabel map[string]Value
	byKey   map[string]Value
}

func New(ticker string) *Stats {
	return &Stats{
		Ticker:      strings.TrimSpace(strings.ToUpper(ticker)),
		RetrievedAt: time.Now(),
		byLabel:     map[string]Value{},
		byKey:       map[string]Value{},
	}
}

func (s *Stats) Set(label string, rawValue string) {
	if s == nil {
		return
	}
	label = strings.TrimSpace(label)
	if label == "" {
		return
	}
	v := ParseValue(rawValue)
	s.byLabel[label] = v
	s.byKey[ToKey(label)] = v
}

func (s *Stats) Get(labelOrKey string) (Value, bool) {
	if s == nil {
		return Value{}, false
	}
	labelOrKey = strings.TrimSpace(labelOrKey)
	if labelOrKey == "" {
		return Value{}, false
	}
	if v, ok := s.byLabel[labelOrKey]; ok {
		return v, true
	}
	if v, ok := s.byKey[ToKey(labelOrKey)]; ok {
		return v, true
	}
	return Value{}, false
}

func (s *Stats) Len() int {
	if s == nil {
		return 0
	}
	return len(s.byLabel)
}

func (s *Stats) Raw(labelOrKey string) (string, bool) {
	v, ok := s.Get(labelOrKey)
	if !ok {
		return "", false
	}
	return v.Raw, true
}

// ToKey normalizes a label to a stable identifier.
// Examples: "P/E" -> "p_e", "52W High" -> "52w_high".
func ToKey(in string) string {
	in = strings.TrimSpace(strings.ToLower(in))
	if in == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(in))
	lastUnderscore := false
	for _, r := range in {
		isAZ := r >= 'a' && r <= 'z'
		is09 := r >= '0' && r <= '9'
		if isAZ || is09 {
			b.WriteRune(r)
			lastUnderscore = false
			continue
		}
		if !lastUnderscore {
			b.WriteByte('_')
			lastUnderscore = true
		}
	}

	out := b.String()
	out = strings.Trim(out, "_")
	return out
}
