package format

import (
	"fmt"
	"sort"
	"strings"
)

// A preset is a base preset plus overrides; risky lists the values its
// style guide asks of risky settings, applied only when risky settings
// are allowed.
type preset struct {
	base      string
	values    map[string]any
	risky     map[string]any
	reference string
}

// keepAll is the neutral layout: every rule leaves the code as written.
var keepAll = map[string]any{
	"indent_style": "keep", "indent_size": 4, "tab_width": 4, "continuation_indent": "keep", "line_ending": "keep",
	"final_newline": false, "trim_trailing_whitespace": false, "remove_bom": false,
	"closing_tag": "keep", "open_tag_own_line": false, "blank_line_after_open_tag": false,
	"max_blank_lines": 0, "line_length": 0, "wrap": "off", "join_fitting_lists": false, "report_long_lines": false,
	"class_brace": "keep", "function_brace": "keep", "closure_brace": "keep", "anonymous_class_brace": "keep", "control_brace": "keep",
	"else_position": "keep", "catch_position": "keep", "empty_body": "keep", "single_line_closures": true,
	"statement_per_line": false, "single_line_throw": false, "blank_line_after_namespace": false, "blank_line_after_imports": false,
	"blank_line_between_import_groups": false, "class_open_blank_line": "keep", "class_close_blank_line": "keep",
	"trim_block_blank_lines": false, "method_separation": "keep", "blank_line_after_trait_use": false, "property_separation": "keep",
	"constant_separation": "keep", "blank_line_before": []string{},
	"spacing": false, "concat_space": "keep", "binary_operator_space": "at_least_one", "cast_space": "keep",
	"not_operator_space": "none", "arrow_function_space": "keep", "closure_space": "one",
	"keyword_case": "keep", "constant_case": "keep", "type_case": "keep", "magic_constant_case": "keep",
	"short_cast": false, "trailing_comma": []string{}, "single_line_trailing_comma": true,
	"array_syntax": "keep", "list_syntax": "keep", "elseif": false, "new_parentheses": "keep",
	"standardize_not_equals": false, "hash_comment": false, "operator_linebreak": "keep",
	"import_order": "none", "import_groups": "none", "no_leading_import_slash": false,
	"single_import_per_statement": false,
	"yoda":                        "keep", "single_quote": false, "remove_unused_imports": false, "declare_strict_types": false,
	"increment_style": "keep", "heredoc_to_nowdoc": false,
}

var presets = map[string]preset{
	// PSR-1: tags and encoding only.
	"psr1": {
		values:    map[string]any{"remove_bom": true},
		reference: "https://www.php-fig.org/psr/psr-1/",
	},
	// PSR-12 (extended coding style), on top of PSR-1.
	"psr12": {
		base: "psr1",
		values: map[string]any{
			"indent_style": "space", "indent_size": 4, "continuation_indent": "one", "line_ending": "lf", "final_newline": true,
			"trim_trailing_whitespace": true, "closing_tag": "remove", "open_tag_own_line": true,
			"blank_line_after_open_tag": true, "max_blank_lines": 1, "line_length": 120, "wrap": "long",
			"class_brace": "next_line", "function_brace": "next_line", "closure_brace": "same_line", "anonymous_class_brace": "same_line",
			"control_brace": "same_line", "else_position": "same_line", "catch_position": "same_line", "single_line_closures": false,
			"statement_per_line": true, "blank_line_after_namespace": true, "blank_line_after_imports": true,
			"blank_line_between_import_groups": true, "class_open_blank_line": "none", "class_close_blank_line": "none", "blank_line_after_trait_use": true,
			"spacing": true, "concat_space": "one", "binary_operator_space": "at_least_one",
			"keyword_case": "lower", "constant_case": "lower", "type_case": "lower", "short_cast": true,
			"elseif": true, "new_parentheses": "add", "import_groups": "class_function_const",
			"no_leading_import_slash": true, "single_import_per_statement": true,
		},
		reference: "https://www.php-fig.org/psr/psr-12/",
	},
	// PER Coding Style 3.1, which extends PSR-12.
	"per": {
		base: "psr12",
		values: map[string]any{
			"empty_body": "inline", "cast_space": "one", "arrow_function_space": "none",
			"trailing_comma":             []string{"arguments", "arrays", "closure_use", "destructuring", "match", "parameters"},
			"single_line_trailing_comma": false, "array_syntax": "short", "operator_linebreak": "start",
		},
		reference: "https://www.php-fig.org/per/coding-style/",
	},
	// Symfony coding standards (the @Symfony rule set builds on PER).
	"symfony": {
		base: "per",
		values: map[string]any{
			"empty_body": "keep", "concat_space": "none", "binary_operator_space": "one",
			"arrow_function_space": "one", "magic_constant_case": "upper", "single_line_closures": true,
			"trailing_comma": []string{"arrays", "destructuring", "match", "parameters"},
			"list_syntax":    "short", "standardize_not_equals": true, "hash_comment": true,
			"operator_linebreak": "start_booleans", "import_order": "alpha", "trim_block_blank_lines": true,
			"method_separation": "one", "blank_line_before": []string{"return"}, "single_line_throw": true,
		},
		risky: map[string]any{
			"yoda": "yoda", "single_quote": true, "remove_unused_imports": true, "increment_style": "pre",
		},
		reference: "https://symfony.com/doc/current/contributing/code/standards.html",
	},
	// Laravel (the coding style Laravel Pint applies by default).
	"laravel": {
		base: "per",
		values: map[string]any{
			"wrap": "off", "concat_space": "none", "binary_operator_space": "one", "not_operator_space": "one",
			"arrow_function_space": "one", "magic_constant_case": "upper", "trailing_comma": []string{"arrays"},
			"list_syntax": "short", "new_parentheses": "remove", "anonymous_class_brace": "next_line", "standardize_not_equals": true,
			"hash_comment": true, "operator_linebreak": "keep", "import_order": "alpha",
			"import_groups": "const_class_function", "trim_block_blank_lines": true, "method_separation": "one",
			"property_separation": "one", "constant_separation": "one", "blank_line_before": []string{"continue", "return"},
		},
		risky: map[string]any{
			"yoda": "regular", "single_quote": true, "remove_unused_imports": true, "increment_style": "post", "heredoc_to_nowdoc": true,
		},
		reference: "https://laravel.com/docs/pint",
	},
	// Drupal coding standards.
	"drupal": {
		base: "per",
		values: map[string]any{
			"indent_size": 2, "line_length": 80, "class_brace": "same_line", "function_brace": "same_line",
			"else_position": "next_line", "catch_position": "next_line", "empty_body": "keep",
			"class_open_blank_line": "one", "class_close_blank_line": "one", "binary_operator_space": "one",
			"constant_case": "upper", "arrow_function_space": "keep", "trailing_comma": []string{"arrays", "parameters"},
			"standardize_not_equals": true, "operator_linebreak": "keep", "import_groups": "none",
			"method_separation": "one",
		},
		reference: "https://project.pages.drupalcode.org/coding_standards/php/coding/",
	},
}

// DefaultPreset is the preset used when nothing names one.
const DefaultPreset = "per"

// Presets returns the preset names, sorted.
func Presets() []string {
	names := make([]string, 0, len(presets))
	for n := range presets {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// PresetReference returns the style guide a preset follows.
func PresetReference(name string) string { return presets[name].reference }

// Layer is one source of setting values, applied over the previous ones.
type Layer struct {
	Source string         // shown by --show-config, e.g. "custos.json"
	Values map[string]any // setting ID → value
}

// Resolved are settings plus where each value came from.
type Resolved struct {
	Settings
	Preset string
	Source map[string]string // setting ID → source of its value
}

// Resolve builds the settings of preset name, then applies layers in
// order. Risky settings keep their neutral value unless allowRisky is
// set, in which case the preset's risky values apply first; a layer may
// always turn a risky setting on explicitly.
func Resolve(name string, allowRisky bool, layers ...Layer) (*Resolved, error) {
	if _, ok := presets[name]; !ok {
		return nil, fmt.Errorf("unknown preset %q (want %s)", name, strings.Join(Presets(), ", "))
	}
	r := &Resolved{Preset: name, Source: map[string]string{}}
	apply := func(values map[string]any, source string) error {
		ids := make([]string, 0, len(values))
		for id := range values {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, id := range ids {
			d, ok := registryByID[id]
			if !ok {
				return fmt.Errorf("%s: unknown format setting %q", source, id)
			}
			if err := d.Set(&r.Settings, values[id]); err != nil {
				return fmt.Errorf("%s: %w", source, err)
			}
			r.Source[id] = source
		}
		return nil
	}
	_ = apply(keepAll, "default")
	var chain []string
	for n := name; n != ""; n = presets[n].base {
		chain = append([]string{n}, chain...)
	}
	for _, n := range chain {
		_ = apply(presets[n].values, "preset "+n)
	}
	if allowRisky {
		for _, n := range chain {
			_ = apply(presets[n].risky, "preset "+n+" (risky)")
		}
	}
	for _, l := range layers {
		if err := apply(l.Values, l.Source); err != nil {
			return nil, err
		}
	}
	return r, nil
}
