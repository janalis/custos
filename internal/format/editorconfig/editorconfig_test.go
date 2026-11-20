package editorconfig

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	src := "\uFEFFroot = TRUE\n" +
		"# comment\n" +
		"indent_style = tab\n" + // preamble pairs other than root are ignored
		"\n" +
		"; another comment\n" +
		"[*.php]\r\n" +
		"  Indent_Size = 4 \r\n" +
		"key = value ; not a comment # either\n" +
		"empty =\n" +
		"no equals sign\n" +
		" = no key\n" +
		"[\n" +
		"[lib/**]\n" +
		"a = b = c\n"
	got := Parse([]byte(src))
	want := &File{
		Root: true,
		Sections: []Section{
			{Glob: "*.php", Pairs: []Pair{
				{"indent_size", "4"},
				{"key", "value ; not a comment # either"},
				{"empty", ""},
			}},
			{Glob: "lib/**", Pairs: []Pair{{"a", "b = c"}}},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Parse:\n got %+v\nwant %+v", got, want)
	}
	if Parse([]byte("root = false\n")).Root {
		t.Error("root = false parsed as root")
	}
	if Parse([]byte("[*]\nroot = true\n")).Root {
		t.Error("root inside a section honored")
	}
}

// fsys returns a reader over files keyed by slash-separated paths.
func fsys(files map[string]string) func(string) ([]byte, error) {
	return func(name string) ([]byte, error) {
		if s, ok := files[filepath.ToSlash(name)]; ok {
			return []byte(s), nil
		}
		return nil, fmt.Errorf("open %s: %w", name, fs.ErrNotExist)
	}
}

func TestResolve(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string
		path  string
		want  Properties
	}{
		{
			name: "no files",
			path: "/p/a.php",
			want: Properties{},
		},
		{
			name: "later sections override earlier ones",
			files: map[string]string{
				"/p/.editorconfig": "[*]\nindent_size = 2\nend_of_line = LF\n[*.php]\nindent_size = 4\n[*.js]\nindent_size = 8\n",
			},
			path: "/p/src/a.php",
			want: Properties{"indent_size": "4", "tab_width": "4", "end_of_line": "lf"},
		},
		{
			name: "nearer files override farther ones; root stops the search",
			files: map[string]string{
				"/.editorconfig":       "[*]\ncustom = Far\n",
				"/p/.editorconfig":     "root = true\n[*]\nindent_style = space\nindent_size = 2\ncustom = Mid\n",
				"/p/src/.editorconfig": "[*.php]\nindent_size = 4\n",
			},
			path: "/p/src/a.php",
			want: Properties{"indent_style": "space", "indent_size": "4", "tab_width": "4", "custom": "Mid"},
		},
		{
			name: "search reaches the filesystem root without root = true",
			files: map[string]string{
				"/.editorconfig":   "[*]\nfar = 1\nnear = 1\n",
				"/p/.editorconfig": "[*]\nnear = 2\n",
			},
			path: "/p/a.php",
			want: Properties{"far": "1", "near": "2"},
		},
		{
			name: "anchored globs are relative to their file's directory",
			files: map[string]string{
				"/.editorconfig":   "[p/src/*.php]\nouter = yes\n[src/*.php]\nwrong = yes\n",
				"/p/.editorconfig": "[src/*.php]\ninner = yes\n[/a.php]\nwrong = yes\n",
			},
			path: "/p/src/a.php",
			want: Properties{"outer": "yes", "inner": "yes"},
		},
		{
			name: "unset removes a key from a farther file",
			files: map[string]string{
				"/.editorconfig":   "[*]\nindent_size = 4\ncharset = utf-8\nother = x\n",
				"/p/.editorconfig": "[*]\nindent_size = UNSET\nother = Unset\n",
			},
			path: "/p/a.php",
			want: Properties{"charset": "utf-8"},
		},
		{
			name:  "unknown values keep their case",
			files: map[string]string{"/.editorconfig": "[*]\nMy_Key = Some Value\nINDENT_STYLE = Space\n"},
			path:  "/a.php",
			want:  Properties{"my_key": "Some Value", "indent_style": "space"},
		},
		{
			name:  "indent_style tab implies indent_size tab",
			files: map[string]string{"/.editorconfig": "[*]\nindent_style = tab\n"},
			path:  "/a.php",
			want:  Properties{"indent_style": "tab", "indent_size": "tab"},
		},
		{
			name:  "indent_size tab takes tab_width",
			files: map[string]string{"/.editorconfig": "[*]\nindent_style = tab\ntab_width = 8\n"},
			path:  "/a.php",
			want:  Properties{"indent_style": "tab", "indent_size": "8", "tab_width": "8"},
		},
		{
			name:  "explicit tab_width is kept",
			files: map[string]string{"/.editorconfig": "[*]\nindent_size = 2\ntab_width = 4\n"},
			path:  "/a.php",
			want:  Properties{"indent_size": "2", "tab_width": "4"},
		},
		{
			name:  "non-numeric indent_size does not set tab_width",
			files: map[string]string{"/.editorconfig": "[*]\nindent_size = wide\n"},
			path:  "/a.php",
			want:  Properties{"indent_size": "wide"},
		},
		{
			name:  "slash-separated relative path stops at the current directory",
			files: map[string]string{"p/.editorconfig": "[*.php]\nx = 1\n", ".editorconfig": "[p/a.php]\ny = 2\n"},
			path:  "p/a.php",
			want:  Properties{"x": "1", "y": "2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Resolve(tt.path, fsys(tt.files))
			if err != nil {
				t.Fatal(err)
			}
			if !maps.Equal(got, tt.want) {
				t.Errorf("Resolve(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestResolveReadError(t *testing.T) {
	boom := errors.New("boom")
	got, err := Resolve("/p/a.php", func(name string) ([]byte, error) {
		if filepath.ToSlash(name) == "/.editorconfig" {
			return nil, boom
		}
		return nil, fs.ErrNotExist
	})
	if !errors.Is(err, boom) || got != nil {
		t.Errorf("Resolve = %v, %v; want nil, %v", got, err, boom)
	}
}

func ptr[T any](v T) *T { return &v }

func TestSettings(t *testing.T) {
	tests := []struct {
		name  string
		props Properties
		want  Settings
	}{
		{name: "empty", props: Properties{}, want: Settings{}},
		{
			name: "all set",
			props: Properties{
				"indent_style": "space", "indent_size": "4", "tab_width": "8", "end_of_line": "crlf",
				"insert_final_newline": "true", "trim_trailing_whitespace": "false",
				"max_line_length": "120", "charset": "utf-8-bom",
			},
			want: Settings{
				IndentStyle: ptr("space"), IndentSize: ptr(4), TabWidth: ptr(8), EndOfLine: ptr("crlf"),
				InsertFinalNewline: ptr(true), TrimTrailingWhitespace: ptr(false),
				MaxLineLength: ptr(120), Charset: ptr("utf-8-bom"),
			},
		},
		{
			name:  "indent_size tab uses tab_width",
			props: Properties{"indent_style": "tab", "indent_size": "tab", "tab_width": "3"},
			want:  Settings{IndentStyle: ptr("tab"), IndentSize: ptr(3), TabWidth: ptr(3)},
		},
		{
			name:  "indent_size tab without tab_width",
			props: Properties{"indent_size": "tab"},
			want:  Settings{},
		},
		{
			name: "invalid values",
			props: Properties{
				"indent_style": "both", "indent_size": "0", "tab_width": "-2", "end_of_line": "nl",
				"insert_final_newline": "yes", "trim_trailing_whitespace": "1",
				"max_line_length": "off", "charset": "ascii",
			},
			want: Settings{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.props.Settings(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Settings() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
