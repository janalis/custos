package security

import (
	"encoding/json"
	"path/filepath"
	"strconv"
	"strings"

	"custos/internal/analysis"
	"custos/internal/syntax"
)

// securityAdvisories inspects composer.json manifests: development packages
// under `require`, and the advisory meta-package in `require-dev`.
type securityAdvisories struct{}

func init() { register(securityAdvisories{}) }

const (
	advisoryPackage = "roave/security-advisories"
	checkerPackage  = "sensiolabs/security-checker"
)

// advisoriesDevDefaults is the default development-package list.
var advisoriesDevDefaults = []string{
	"phpunit/phpunit", "johnkary/phpunit-speedtrap", "brianium/paratest", "mybuilder/phpunit-accelerator",
	"codedungeon/phpunit-result-printer", "spatie/phpunit-watcher", "symfony/phpunit-bridge", "symfony/debug",
	"symfony/maker-bundle", "zendframework/zend-test", "zendframework/zend-debug", "yiisoft/yii2-gii",
	"yiisoft/yii2-debug", "orchestra/testbench", "barryvdh/laravel-debugbar", "codeception/codeception",
	"behat/behat", "yiisoft/yii2-coding-standards", "phpmd/phpmd", "sebastian/phpcpd", "phan/phan",
	"wimg/php-compatibility", "phing/phing", "roave/security-advisories",
	"phpspec/prophecy", "phpspec/phpspec", "humbug/humbug", "infection/infection", "mockery/mockery",
	"satooshi/php-coveralls", "mikey179/vfsStream", "filp/whoops", "friendsofphp/php-cs-fixer",
	"phpstan/phpstan", "vimeo/psalm", "jakub-onderka/php-parallel-lint", "squizlabs/php_codesniffer",
	"slevomat/coding-standard", "doctrine/coding-standard", "phpcompatibility/php-compatibility",
	"zendframework/zend-coding-standard", "wp-coding-standards/wpcs", "pdepend/pdepend", "povils/phpmnd",
	"phpro/grumphp", "sstalle/php7cc", "composer/composer", "kalessil/production-dependencies-guard",
}

func (securityAdvisories) ID() string { return "SecurityAdvisories" }

func (securityAdvisories) Kinds() []syntax.NodeKind { return nil }

func (securityAdvisories) Check(*analysis.Context, syntax.Node) {}

// FilePatterns makes the runner discover composer manifests.
func (securityAdvisories) FilePatterns() []string { return []string{"composer.json"} }

func (securityAdvisories) CheckFile(ctx *analysis.Context) {
	if filepath.Base(ctx.File.Path) != "composer.json" { // D0
		return
	}
	root, ok := parseJSON(ctx.Src)
	if !ok || root.kind != jsonObject {
		return
	}
	dev := map[string]bool{}
	// Composer package names are case-insensitive: DEV is compared lower-cased.
	for _, d := range advisoriesDevDefaults {
		dev[strings.ToLower(d)] = true
	}
	for _, d := range ctx.List("optionConfiguration") {
		dev[strings.ToLower(d)] = true
	}
	// D1; a metapackage only bundles requirements meant to be pulled in
	// elsewhere (often into require-dev): custos diverges.
	if t := root.get("type"); t != nil && t.kind == jsonString && (t.str == "library" || t.str == "metapackage") {
		return
	}
	owner := "" // D2
	if nm := root.get("name"); nm != nil && nm.kind == jsonString {
		switch {
		case dev[strings.ToLower(nm.str)]:
			owner = nm.str
		case strings.Contains(nm.str, "/"):
			owner = nm.str[:strings.IndexByte(nm.str, '/')+1]
		}
	}
	if owner != "" && dev[strings.ToLower(owner)] {
		return
	}
	rqMember := root.member("require") // D3
	if rqMember == nil || rqMember.value.kind != jsonObject {
		return
	}
	thirdParty, secured := false, false
	for _, m := range rqMember.value.members {
		if m.value.kind != jsonString {
			continue
		}
		p, v := strings.ToLower(m.key), strings.ToLower(m.value.str)
		if p == "" || v == "" {
			continue
		}
		if ctx.Bool("REPORT_MISPLACED_DEPENDENCIES") && dev[p] { // D4
			ctx.Report(m.keySpan, "Development package in require; move it to require-dev.")
		}
		if strings.Contains(p, "/") && (owner == "" || !strings.HasPrefix(p, strings.ToLower(owner))) { // D5
			thirdParty = true
		}
		if p == checkerPackage { // D6
			secured = true
		}
	}
	if !ctx.Bool("REPORT_MISSING_ROAVE_ADVISORIES") {
		return
	}
	devMember := root.member("require-dev") // D7
	if devMember != nil && devMember.value.kind == jsonObject {
		for _, m := range devMember.value.members {
			if m.value.kind != jsonString {
				continue
			}
			p, v := strings.ToLower(m.key), strings.ToLower(m.value.str)
			if p == "" || v == "" {
				continue
			}
			if p == checkerPackage {
				secured = true
			}
			if p == advisoryPackage {
				secured = true
				if v != "dev-latest" {
					ctx.Report(m.value.span, "Constrain the advisory package to dev-latest.")
				}
				break
			}
		}
	}
	if secured || !thirdParty { // D8
		return
	}
	pair := strconv.Quote(advisoryPackage) + `: "dev-latest"`
	var fixes []analysis.Fix
	switch {
	case devMember == nil: // F1
		at := rqMember.value.span.End
		text := "\n  ,\n  \"require-dev\": {\n    " + pair + "\n  }"
		fixes = append(fixes, advisoriesFix(at, text))
	case devMember.value.kind == jsonObject: // F2
		at := devMember.value.span.Start + 1
		text := "\n    " + pair
		if len(devMember.value.members) > 0 {
			text += "\n    ,"
		}
		fixes = append(fixes, advisoriesFix(at, text))
	}
	ctx.Report(rqMember.keySpan, "Add the security advisories package (dev-latest) to require-dev.", fixes...)
}

func advisoriesFix(at uint32, text string) analysis.Fix {
	return analysis.Fix{Title: "Add the advisory package to require-dev", Edits: func() []analysis.TextEdit {
		return []analysis.TextEdit{{Span: syntax.Span{Start: at, End: at}, NewText: text}}
	}}
}

// ---- minimal JSON parser with byte spans --------------------------------------------

type jsonKind uint8

const (
	jsonNull jsonKind = iota
	jsonBool
	jsonNumber
	jsonString
	jsonArray
	jsonObject
)

type jsonValue struct {
	kind    jsonKind
	span    syntax.Span
	str     string // decoded string value
	members []jsonMember
	items   []*jsonValue
}

type jsonMember struct {
	key     string
	keySpan syntax.Span
	value   *jsonValue
}

func (v *jsonValue) member(key string) *jsonMember {
	for i := range v.members {
		if v.members[i].key == key {
			return &v.members[i]
		}
	}
	return nil
}

func (v *jsonValue) get(key string) *jsonValue {
	if m := v.member(key); m != nil {
		return m.value
	}
	return nil
}

type jsonParser struct {
	src []byte
	pos int
}

// parseJSON parses a complete JSON document; ok is false on syntax errors.
func parseJSON(src []byte) (*jsonValue, bool) {
	p := &jsonParser{src: src}
	v, ok := p.value(0)
	if !ok {
		return nil, false
	}
	p.ws()
	return v, p.pos == len(src)
}

func (p *jsonParser) ws() {
	for p.pos < len(p.src) {
		switch p.src[p.pos] {
		case ' ', '\t', '\n', '\r':
			p.pos++
		default:
			return
		}
	}
}

func (p *jsonParser) value(depth int) (*jsonValue, bool) {
	if depth > 512 {
		return nil, false
	}
	p.ws()
	if p.pos >= len(p.src) {
		return nil, false
	}
	start := p.pos
	switch c := p.src[p.pos]; {
	case c == '{':
		p.pos++
		v := &jsonValue{kind: jsonObject}
		p.ws()
		if p.pos < len(p.src) && p.src[p.pos] == '}' {
			p.pos++
			v.span = syntax.Span{Start: uint32(start), End: uint32(p.pos)}
			return v, true
		}
		for {
			p.ws()
			ks := p.pos
			key, ok := p.str()
			if !ok {
				return nil, false
			}
			keySpan := syntax.Span{Start: uint32(ks), End: uint32(p.pos)}
			p.ws()
			if p.pos >= len(p.src) || p.src[p.pos] != ':' {
				return nil, false
			}
			p.pos++
			val, ok := p.value(depth + 1)
			if !ok {
				return nil, false
			}
			v.members = append(v.members, jsonMember{key: key, keySpan: keySpan, value: val})
			p.ws()
			if p.pos >= len(p.src) {
				return nil, false
			}
			if p.src[p.pos] == ',' {
				p.pos++
				continue
			}
			if p.src[p.pos] != '}' {
				return nil, false
			}
			p.pos++
			v.span = syntax.Span{Start: uint32(start), End: uint32(p.pos)}
			return v, true
		}
	case c == '[':
		p.pos++
		v := &jsonValue{kind: jsonArray}
		p.ws()
		if p.pos < len(p.src) && p.src[p.pos] == ']' {
			p.pos++
			v.span = syntax.Span{Start: uint32(start), End: uint32(p.pos)}
			return v, true
		}
		for {
			it, ok := p.value(depth + 1)
			if !ok {
				return nil, false
			}
			v.items = append(v.items, it)
			p.ws()
			if p.pos >= len(p.src) {
				return nil, false
			}
			if p.src[p.pos] == ',' {
				p.pos++
				continue
			}
			if p.src[p.pos] != ']' {
				return nil, false
			}
			p.pos++
			v.span = syntax.Span{Start: uint32(start), End: uint32(p.pos)}
			return v, true
		}
	case c == '"':
		s, ok := p.str()
		if !ok {
			return nil, false
		}
		return &jsonValue{kind: jsonString, str: s, span: syntax.Span{Start: uint32(start), End: uint32(p.pos)}}, true
	default:
		for _, lit := range []struct {
			text string
			kind jsonKind
		}{{"true", jsonBool}, {"false", jsonBool}, {"null", jsonNull}} {
			if strings.HasPrefix(string(p.src[p.pos:min(len(p.src), p.pos+5)]), lit.text) {
				p.pos += len(lit.text)
				return &jsonValue{kind: lit.kind, span: syntax.Span{Start: uint32(start), End: uint32(p.pos)}}, true
			}
		}
		for p.pos < len(p.src) && strings.IndexByte("+-0123456789.eE", p.src[p.pos]) >= 0 {
			p.pos++
		}
		if p.pos == start {
			return nil, false
		}
		return &jsonValue{kind: jsonNumber, span: syntax.Span{Start: uint32(start), End: uint32(p.pos)}}, true
	}
}

// str parses a JSON string at pos and returns its decoded value.
func (p *jsonParser) str() (string, bool) {
	if p.pos >= len(p.src) || p.src[p.pos] != '"' {
		return "", false
	}
	start := p.pos
	p.pos++
	escaped := false
	for p.pos < len(p.src) {
		c := p.src[p.pos]
		switch {
		case c == '\\':
			escaped = true
			p.pos += 2
			continue
		case c == '"':
			p.pos++
			raw := string(p.src[start:p.pos])
			if !escaped {
				return raw[1 : len(raw)-1], true
			}
			// JSON escapes (`\/`, `\u00e9`…); an invalid escape keeps
			// the raw contents.
			var s string
			if json.Unmarshal([]byte(raw), &s) != nil {
				return raw[1 : len(raw)-1], true
			}
			return s, true
		}
		p.pos++
	}
	return "", false
}
