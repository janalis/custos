//go:build unix

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
)

// parsers count the findings in a tool's standard output. They are not
// comparable across tools: each reports what its own rules look for.
var parsers = map[string]func([]byte) (int, error){
	"custos":       custosFindings,
	"custos-fix":   custosFixEdits,
	"phpstan":      phpstanFindings,
	"psalm":        jsonArrayLen,
	"phan":         jsonArrayLen,
	"mago":         magoFindings,
	"phpcs":        phpcsFindings,
	"php-cs-fixer": phpCsFixerFindings,
	"rector":       rectorFindings,
}

// decode unmarshals the JSON document of out, skipping any text a tool
// prints before it (banners, deprecation notices).
func decode(out []byte, v any) error {
	i := bytes.IndexAny(out, "{[")
	if i < 0 {
		return errors.New("no JSON in output")
	}
	return json.NewDecoder(bytes.NewReader(out[i:])).Decode(v)
}

func custosFindings(out []byte) (int, error) {
	var r struct {
		Findings []json.RawMessage `json:"findings"`
	}
	err := decode(out, &r)
	return len(r.Findings), err
}

var fixEdits = regexp.MustCompile(`would fix \d+ file\(s\), (\d+) edit\(s\)`)

// custosFixEdits counts the edits `custos fix -dry-run` would apply.
func custosFixEdits(out []byte) (int, error) {
	m := fixEdits.FindSubmatch(out)
	if m == nil {
		if bytes.Contains(out, []byte("nothing to fix")) {
			return 0, nil
		}
		return 0, errors.New("no fix summary in output")
	}
	return strconv.Atoi(string(m[1]))
}

func phpstanFindings(out []byte) (int, error) {
	var r struct {
		Totals *struct {
			Errors     int `json:"errors"`
			FileErrors int `json:"file_errors"`
		} `json:"totals"`
	}
	if err := decode(out, &r); err != nil {
		return 0, err
	}
	if r.Totals == nil {
		return 0, errors.New("no totals in output")
	}
	return r.Totals.Errors + r.Totals.FileErrors, nil
}

func jsonArrayLen(out []byte) (int, error) {
	var r []json.RawMessage
	err := decode(out, &r)
	return len(r), err
}

func magoFindings(out []byte) (int, error) {
	var r struct {
		Issues *[]json.RawMessage `json:"issues"`
	}
	if err := decode(out, &r); err != nil {
		return 0, err
	}
	if r.Issues == nil {
		return 0, errors.New("no issues in output")
	}
	return len(*r.Issues), nil
}

func phpcsFindings(out []byte) (int, error) {
	var r struct {
		Totals *struct {
			Errors   int `json:"errors"`
			Warnings int `json:"warnings"`
		} `json:"totals"`
	}
	if err := decode(out, &r); err != nil {
		return 0, err
	}
	if r.Totals == nil {
		return 0, errors.New("no totals in output")
	}
	return r.Totals.Errors + r.Totals.Warnings, nil
}

// phpCsFixerFindings counts the files php-cs-fixer would change.
func phpCsFixerFindings(out []byte) (int, error) {
	var r struct {
		Files *[]json.RawMessage `json:"files"`
	}
	if err := decode(out, &r); err != nil {
		return 0, err
	}
	if r.Files == nil {
		return 0, errors.New("no files in output")
	}
	return len(*r.Files), nil
}

// rectorFindings counts the files Rector would change.
func rectorFindings(out []byte) (int, error) {
	var r struct {
		Totals *struct {
			ChangedFiles int `json:"changed_files"`
		} `json:"totals"`
	}
	if err := decode(out, &r); err != nil {
		return 0, err
	}
	if r.Totals == nil {
		return 0, errors.New("no totals in output")
	}
	return r.Totals.ChangedFiles, nil
}
