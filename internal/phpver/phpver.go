// Package phpver models target PHP language versions (5.3 – 8.5).
package phpver

import (
	"fmt"
	"strconv"
	"strings"
)

// Version is major*100 + minor, e.g. 803 for PHP 8.3.
type Version uint16

const (
	PHP53 Version = 503
	PHP54 Version = 504
	PHP55 Version = 505
	PHP56 Version = 506
	PHP70 Version = 700
	PHP71 Version = 701
	PHP72 Version = 702
	PHP73 Version = 703
	PHP74 Version = 704
	PHP80 Version = 800
	PHP81 Version = 801
	PHP82 Version = 802
	PHP83 Version = 803
	PHP84 Version = 804
	PHP85 Version = 805

	Min     = PHP53
	Max     = PHP85
	Default = PHP84
)

// AtLeast reports whether v >= o.
func (v Version) AtLeast(o Version) bool { return v >= o }

// Below reports whether v < o.
func (v Version) Below(o Version) bool { return v < o }

func (v Version) String() string { return fmt.Sprintf("%d.%d", v/100, v%100) }

// Parse accepts "8.3", "8.3.1", "80300" style is not supported; "" yields Default.
func Parse(s string) (Version, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Default, nil
	}
	parts := strings.SplitN(s, ".", 3)
	if len(parts) < 2 {
		return 0, fmt.Errorf("phpver: invalid version %q", s)
	}
	maj, err1 := strconv.Atoi(parts[0])
	min, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil || maj < 5 || maj > 8 || min < 0 || min > 99 {
		return 0, fmt.Errorf("phpver: invalid version %q", s)
	}
	v := Version(maj*100 + min)
	if v < Min || v > Max || (maj == 5 && min > 6) || maj == 6 || (maj == 7 && min > 4) {
		return 0, fmt.Errorf("phpver: unsupported version %s (supported %s–%s)", v, Min, Max)
	}
	return v, nil
}

// MustParse is Parse that panics; for tests and constants.
func MustParse(s string) Version {
	v, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return v
}
