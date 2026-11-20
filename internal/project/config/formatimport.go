package config

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/janalis/custos/internal/format"
	"github.com/janalis/custos/internal/php/syntax"
)

// toolFormat is the format configuration read from another tool: the preset
// closest to its style and the settings it states, as layers named after its
// configuration file.
type toolFormat struct {
	preset, source string
	layers         []format.Layer
}

// importFormat reads the style of the first formatter configuration found at
// the root: Laravel Pint, then PHP-CS-Fixer, then Mago. Without any, the
// default preset applies. Only what maps unambiguously onto custos settings
// is imported; the tools' other rules are left to the preset.
func importFormat(read func(string) ([]byte, error)) toolFormat {
	if read == nil {
		return toolFormat{preset: format.DefaultPreset, source: "default"}
	}
	if t, ok := pintFormat(read); ok {
		return t
	}
	if t, ok := csFixerFormat(read); ok {
		return t
	}
	if t, ok := magoFormat(read); ok {
		return t
	}
	return toolFormat{preset: format.DefaultPreset, source: "default"}
}

// pintPresets maps Pint presets onto custos presets; "empty" (no rule)
// is closest to psr1, which only fixes tags and encoding.
var pintPresets = map[string]string{
	"laravel": "laravel", "per": "per", "psr12": "psr12", "symfony": "symfony", "empty": "psr1",
}

// pintFormat reads pint.json: its "preset", laravel (Pint's default) when it
// names none or one custos does not know.
func pintFormat(read func(string) ([]byte, error)) (toolFormat, bool) {
	const name = "pint.json"
	b, err := read(name)
	if err != nil {
		return toolFormat{}, false
	}
	var f struct {
		Preset string `json:"preset"`
	}
	_ = json.Unmarshal(b, &f)
	preset, ok := pintPresets[strings.ToLower(f.Preset)]
	if !ok {
		preset = "laravel"
	}
	return toolFormat{preset: preset, source: name}, true
}

// csFixerFormat reads the PHP-CS-Fixer configuration (parsed, never run):
// the last layout rule set it enables (`'@Symfony' => true`) gives the
// preset, `setIndent()` the indentation and `setLineEnding()` the line
// ending, when written as string literals. Risky sets (`@Symfony:risky`)
// leave the layout alone and are ignored. Without a set, PHP-CS-Fixer
// applies @PSR12.
func csFixerFormat(read func(string) ([]byte, error)) (toolFormat, bool) {
	for _, name := range csFixerFiles {
		b, err := read(name)
		if err != nil {
			continue
		}
		t := toolFormat{preset: "psr12", source: name}
		values := map[string]any{}
		f := syntax.ParseBest(name, b, syntax.Options{})
		syntax.InspectFile(f, func(n syntax.Node) bool {
			switch n := n.(type) {
			case *syntax.ArrayItem:
				if preset := csFixerSet(csFixerString(n.Key)); preset != "" && csFixerConst(n.Value, "true") {
					t.preset = preset
				}
			case *syntax.MethodCall:
				csFixerSetting(n, values)
			}
			return true
		})
		if len(values) > 0 {
			t.layers = []format.Layer{{Source: name, Values: values}}
		}
		return t, true
	}
	return toolFormat{}, false
}

// csFixerSet returns the preset closest to a PHP-CS-Fixer rule set: @PSR2
// is the PSR-12 ancestor, @PhpCsFixer extends @Symfony, and @PER-CS sets
// of any version follow PER.
func csFixerSet(set string) string {
	switch set = strings.ToLower(set); {
	case set == "@symfony" || set == "@phpcsfixer":
		return "symfony"
	case set == "@psr12" || set == "@psr2":
		return "psr12"
	case set == "@psr1":
		return "psr1"
	case set == "@per" || strings.HasPrefix(set, "@per-cs") && !strings.HasSuffix(set, ":risky"):
		return "per"
	}
	return ""
}

// csFixerSetting reads `->setIndent('…')` and `->setLineEnding('…')`.
func csFixerSetting(call *syntax.MethodCall, values map[string]any) {
	id, ok := call.Name.(*syntax.Identifier)
	if !ok || call.Args == nil || len(call.Args.Args) == 0 {
		return
	}
	arg, ok := call.Args.Args[0].(*syntax.Arg)
	if !ok {
		return
	}
	text, ok := phpString(arg.Value)
	if !ok {
		return
	}
	switch strings.ToLower(id.Value) {
	case "setindent":
		switch {
		case text != "" && strings.Trim(text, "\t") == "":
			values["indent_style"] = "tab"
		case text != "" && strings.Trim(text, " ") == "":
			values["indent_style"], values["indent_size"] = "space", len(text)
		}
	case "setlineending":
		switch text {
		case "\n":
			values["line_ending"] = "lf"
		case "\r\n":
			values["line_ending"] = "crlf"
		}
	}
}

// phpString returns the value of a quoted string literal without
// interpolation, decoding the escapes that spell whitespace.
func phpString(e syntax.Expr) (string, bool) {
	lit, ok := e.(*syntax.Literal)
	if !ok || lit.LitKind != syntax.LitString || len(lit.Raw) < 2 {
		return "", false
	}
	body := lit.Raw[1 : len(lit.Raw)-1]
	switch lit.Raw[0] {
	case '\'':
		return strings.NewReplacer(`\\`, `\`, `\'`, `'`).Replace(body), true
	case '"':
		return strings.NewReplacer(`\t`, "\t", `\n`, "\n", `\r`, "\r", `\\`, `\`, `\"`, `"`).Replace(body), true
	}
	return "", false
}

// magoPresets maps Mago formatter presets onto custos presets.
var magoPresets = map[string]string{
	"default": "per", "psr-12": "psr12", "psr12": "psr12", "laravel": "laravel", "pint": "laravel", "drupal": "drupal",
}

// magoFormat reads the `[formatter]` table of the Mago configuration:
// `preset` (per when absent or unknown, as Mago's default style follows
// PER), `print-width` as line_length, `tab-width` as indent_size and
// tab_width, `use-tabs` as indent_style. Keys may use `_` for `-`.
// `single-quote` is not imported: single_quote is a risky setting the
// configuration turns on explicitly.
func magoFormat(read func(string) ([]byte, error)) (toolFormat, bool) {
	for _, name := range magoFiles {
		b, err := read(name)
		if err != nil {
			continue
		}
		t := toolFormat{preset: format.DefaultPreset, source: name}
		values := map[string]any{}
		table := ""
		for _, line := range strings.Split(string(b), "\n") {
			line = strings.TrimSpace(magoStripComment(line))
			if strings.HasPrefix(line, "[") {
				table = strings.Trim(line, "[] \t")
				continue
			}
			key, value, ok := strings.Cut(line, "=")
			if table != "formatter" || !ok {
				continue
			}
			key = strings.ReplaceAll(strings.Trim(strings.TrimSpace(key), `"'`), "_", "-")
			value = strings.TrimSpace(value)
			n, isInt := strconv.Atoi(value)
			switch {
			case key == "preset":
				if p, ok := magoPresets[strings.ToLower(strings.Trim(value, `"'`))]; ok {
					t.preset = p
				}
			case key == "print-width" && isInt == nil && n > 0:
				values["line_length"] = n
			case key == "tab-width" && isInt == nil && n > 0:
				values["indent_size"], values["tab_width"] = n, n
			case key == "use-tabs" && (value == "true" || value == "false"):
				values["indent_style"] = map[bool]string{true: "tab", false: "space"}[value == "true"]
			}
		}
		if len(values) > 0 {
			t.layers = []format.Layer{{Source: name, Values: values}}
		}
		return t, true
	}
	return toolFormat{}, false
}
