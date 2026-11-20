package project

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/janalis/custos/internal/format"
	phpversion "github.com/janalis/custos/internal/php/version"
)

// FormatOptions configure FormatFiles.
type FormatOptions struct {
	// Settings returns the settings of a file.
	Settings func(path string) (*format.Resolved, error)
	PHP      phpversion.Version
	// Cache skips the files known to be formatted; nil formats every one.
	Cache *FormatCache
}

// FormattedFile is the outcome of formatting one file. Result.Output is the
// source when formatting changes nothing.
type FormattedFile struct {
	Source []byte
	Result format.Result
	Perm   os.FileMode
	// Cached is set when the cache knew the file formatted.
	Cached bool
	// Err is set when the file could not be formatted; it wraps ErrSkipped
	// for a file that is not regular (never formatted, as fix never fixes
	// one).
	Err error
}

// FormatFiles formats every file with one worker per CPU. Outcomes are in
// input order; adapters own the writes.
func FormatFiles(files []string, opt FormatOptions) []FormattedFile {
	outs := make([]FormattedFile, len(files))
	var prints sync.Map // *format.Resolved → settings fingerprint
	_ = parallel(context.Background(), len(files), 0, func(i int) {
		src, perm, err := readFixable(files[i], nil)
		if err != nil {
			outs[i].Err = err
			return
		}
		s, err := opt.Settings(files[i])
		if err != nil {
			outs[i].Err = fmt.Errorf("%s: %w", files[i], err)
			return
		}
		var key FormatKey
		if opt.Cache != nil {
			fp, ok := prints.Load(s)
			if !ok {
				fp, _ = prints.LoadOrStore(s, fmt.Sprintf("%d\x00%#v", opt.PHP, s.Settings))
			}
			key = opt.Cache.Key(fp.(string), src)
			if opt.Cache.Formatted(key) {
				outs[i] = FormattedFile{Source: src, Result: format.Result{Output: src}, Perm: perm, Cached: true}
				return
			}
		}
		res := formatSource(files[i], src, format.Options{Settings: &s.Settings, PHP: opt.PHP})
		if opt.Cache != nil && !res.Changed() && res.Skipped == "" {
			opt.Cache.Store(key)
		}
		outs[i] = FormattedFile{Source: src, Result: res, Perm: perm}
	}, func(i int, r any) {
		outs[i] = FormattedFile{Err: fmt.Errorf("%s: internal failure, left unchanged: %v", files[i], r)}
	})
	return outs
}

// formatSource is format.Source; tests replace it to simulate a crash.
var formatSource = format.Source
