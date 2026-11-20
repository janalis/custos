package config

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/janalis/custos/internal/format"
	"github.com/janalis/custos/internal/format/editorconfig"
	"github.com/janalis/custos/internal/platform/safeio"
)

// AutoPreset picks the format preset from the project's other tools (see
// importFormat).
const AutoPreset = "auto"

// FormatFile is the JSON shape of the "format" section of custos.json.
type FormatFile struct {
	// Preset names a format preset, or "auto" (the default).
	Preset string `json:"preset,omitempty"`
	// Settings maps setting IDs to values, over the preset.
	Settings map[string]any `json:"settings,omitempty"`
	// Risky lists risky settings to turn on: a boolean one becomes true,
	// another takes the value the preset gives it when risky settings are
	// allowed, else its first allowed value. Settings still wins.
	Risky []string `json:"risky,omitempty"`
	// AllowRisky applies the preset's own risky values.
	AllowRisky *bool `json:"allowRisky,omitempty"`
	// Exclude lists globs of files not to format, relative to the root
	// (EditorConfig glob syntax, `**` included).
	Exclude []string `json:"exclude,omitempty"`
	// EditorConfig reads .editorconfig files (default true).
	EditorConfig *bool `json:"editorconfig,omitempty"`
}

// FormatConfig is the resolved "format" section.
type FormatConfig struct {
	Preset       string // "auto" or a preset name
	Settings     map[string]any
	Risky        []string
	AllowRisky   bool
	Exclude      []string
	EditorConfig bool
	// imported reads the other tools' configuration once, on first use.
	imported func() toolFormat
}

// FormatFlags are the command-line format options, applied over the
// configuration.
type FormatFlags struct {
	Preset     string         // "" or "auto": the configuration's
	Set        map[string]any // setting ID → value
	AllowRisky bool
	// Within confines the .editorconfig search to a directory: files above
	// it count as absent (an MCP workspace reads nothing outside its root).
	Within string
}

// Sources of the format layers.
const (
	sourceConfig      = FileName
	sourceCommandLine = "command line"
	sourceEditor      = ".editorconfig"
)

func resolveFormat(f *FormatFile, read func(string) ([]byte, error)) (FormatConfig, error) {
	if f == nil {
		f = &FormatFile{}
	}
	fc := FormatConfig{Preset: strings.ToLower(f.Preset), Settings: f.Settings, Risky: f.Risky, Exclude: f.Exclude, EditorConfig: true}
	fc.imported = sync.OnceValue(func() toolFormat { return importFormat(read) })
	if fc.Preset == "" {
		fc.Preset = AutoPreset
	}
	if f.AllowRisky != nil {
		fc.AllowRisky = *f.AllowRisky
	}
	if f.EditorConfig != nil {
		fc.EditorConfig = *f.EditorConfig
	}
	if fc.Preset != AutoPreset && !slices.Contains(format.Presets(), fc.Preset) {
		return fc, fmt.Errorf("%s: format: unknown preset %q (want auto, %s)", FileName, f.Preset, strings.Join(format.Presets(), ", "))
	}
	for _, id := range fc.Risky {
		if d, ok := format.Lookup(id); !ok || !d.Risky {
			return fc, fmt.Errorf("%s: format: risky entry %q is not a risky setting", FileName, id)
		}
	}
	if _, err := format.Resolve(format.DefaultPreset, false, format.Layer{Source: sourceConfig, Values: fc.Settings}); err != nil {
		return fc, fmt.Errorf("format: %w", err)
	}
	return fc, nil
}

// FormatResolver resolves the format settings of each file. It is safe for
// concurrent use; the settings it returns are shared and must not be
// modified.
type FormatResolver struct {
	root         string
	preset       string
	presetSource string
	allowRisky   bool
	editorConfig bool
	base, top    []format.Layer
	// read reads an .editorconfig file by OS path.
	read func(string) ([]byte, error)

	mu sync.Mutex
	// dirs caches the .editorconfig file of each directory.
	dirs map[string]*editorDir
	// byKey caches the settings by the EditorConfig sections that apply.
	byKey map[string]*format.Resolved
}

type editorDir struct {
	data []byte
	file *editorconfig.File // nil without a file
	err  error
}

// FormatResolver returns the resolver of the format settings with flags.
// The layers apply in this order, later ones winning: the preset; what the
// project's other tools configure (only when the preset is "auto" in both
// the configuration and flags); the file's .editorconfig (unless disabled);
// custos.json; the command line.
func (c *Config) FormatResolver(flags FormatFlags) (*FormatResolver, error) {
	fc := c.Format
	if fc.imported == nil { // a Config built by hand
		fc, _ = resolveFormat(nil, nil)
	}
	r := &FormatResolver{
		root: c.Root, allowRisky: fc.AllowRisky || flags.AllowRisky, editorConfig: fc.EditorConfig,
		read:  confinedRead(flags.Within),
		dirs:  map[string]*editorDir{},
		byKey: map[string]*format.Resolved{},
	}
	switch preset := strings.ToLower(flags.Preset); {
	case preset != "" && preset != AutoPreset:
		if !slices.Contains(format.Presets(), preset) {
			return nil, fmt.Errorf("unknown preset %q (want auto, %s)", flags.Preset, strings.Join(format.Presets(), ", "))
		}
		r.preset, r.presetSource = preset, sourceCommandLine
	case fc.Preset != AutoPreset:
		r.preset, r.presetSource = fc.Preset, sourceConfig
	default:
		imported := fc.imported()
		r.preset, r.presetSource, r.base = imported.preset, imported.source, imported.layers
	}
	// Risky entries take the preset's risky value; the preset is known.
	risky, _ := format.Resolve(r.preset, true)
	own := map[string]any{}
	for _, id := range fc.Risky {
		d, _ := format.Lookup(id)
		v := d.Get(&risky.Settings)
		switch {
		case d.Kind == format.KindBool:
			v = true
		case v == "keep":
			v = d.Values[0]
		}
		own[id] = v
	}
	maps.Copy(own, fc.Settings)
	r.top = []format.Layer{{Source: sourceConfig, Values: own}, {Source: sourceCommandLine, Values: flags.Set}}
	// Check the layers once: the EditorConfig values are always valid.
	if _, err := format.Resolve(r.preset, r.allowRisky, append(slices.Clip(r.base), r.top...)...); err != nil {
		return nil, err
	}
	return r, nil
}

// FormatSettings returns the format settings of the file path.
func (c *Config) FormatSettings(path string, flags FormatFlags) (*format.Resolved, error) {
	r, err := c.FormatResolver(flags)
	if err != nil {
		return nil, err
	}
	return r.Settings(path)
}

// Preset returns the preset the settings start from and where it was
// chosen: "command line", custos.json, the other tool's configuration file,
// or "default".
func (r *FormatResolver) Preset() (name, source string) { return r.preset, r.presetSource }

// Settings returns the settings of the file path.
func (r *FormatResolver) Settings(path string) (*format.Resolved, error) {
	// Abs fails only without a working directory; the walk then starts
	// from the path as given.
	abs, _ := filepath.Abs(path)
	r.mu.Lock()
	defer r.mu.Unlock()
	key := ""
	if r.editorConfig {
		var err error
		if key, err = r.editorKey(abs); err != nil {
			return nil, err
		}
	}
	if res, ok := r.byKey[key]; ok {
		return res, nil
	}
	layers := slices.Clip(r.base)
	if r.editorConfig {
		// editorKey read every file already: cachedRead cannot fail.
		props, _ := editorconfig.Resolve(abs, r.cachedRead)
		layers = append(layers, format.Layer{Source: sourceEditor, Values: editorValues(props.Settings())})
	}
	// FormatResolver checked the other layers, and EditorConfig values are
	// valid settings.
	res, _ := format.Resolve(r.preset, r.allowRisky, append(layers, r.top...)...)
	r.byKey[key] = res
	return res, nil
}

// editorKey identifies the EditorConfig sections that apply to abs: files
// sharing it share their properties. It walks the directories as
// editorconfig.Resolve does.
func (r *FormatResolver) editorKey(abs string) (string, error) {
	var key strings.Builder
	p := filepath.Clean(abs)
	dir, rel := filepath.Dir(p), filepath.Base(p)
	for {
		e := r.dir(dir)
		if e.err != nil {
			return "", e.err
		}
		if e.file != nil {
			key.WriteString(dir)
			key.WriteByte(0)
			for i, sec := range e.file.Sections {
				if editorconfig.Match(sec.Glob, rel) {
					key.WriteString(strconv.Itoa(i))
					key.WriteByte(',')
				}
			}
			key.WriteByte('\n')
			if e.file.Root {
				break
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		rel = filepath.Base(dir) + "/" + rel
		dir = parent
	}
	return key.String(), nil
}

// dir returns the .editorconfig file of the directory, read once; r.mu is
// held.
func (r *FormatResolver) dir(dir string) *editorDir {
	if e, ok := r.dirs[dir]; ok {
		return e
	}
	e := &editorDir{}
	e.data, e.err = r.read(filepath.Join(dir, ".editorconfig"))
	if errors.Is(e.err, fs.ErrNotExist) {
		e.err = nil
	} else if e.err == nil {
		e.file = editorconfig.Parse(e.data)
	}
	r.dirs[dir] = e
	return e
}

// cachedRead serves editorconfig.Resolve the files dir read; r.mu is held.
func (r *FormatResolver) cachedRead(name string) ([]byte, error) {
	e := r.dir(filepath.Dir(name))
	if e.file == nil {
		return nil, fs.ErrNotExist
	}
	return e.data, nil
}

// editorValues maps EditorConfig properties to format settings.
func editorValues(s editorconfig.Settings) map[string]any {
	v := map[string]any{}
	if s.IndentStyle != nil {
		v["indent_style"] = *s.IndentStyle
	}
	setInt(v, "indent_size", s.IndentSize)
	setInt(v, "tab_width", s.TabWidth)
	if s.EndOfLine != nil {
		v["line_ending"] = *s.EndOfLine
	}
	if s.InsertFinalNewline != nil {
		v["final_newline"] = *s.InsertFinalNewline
	}
	if s.TrimTrailingWhitespace != nil {
		v["trim_trailing_whitespace"] = *s.TrimTrailingWhitespace
	}
	setInt(v, "line_length", s.MaxLineLength)
	return v
}

// setInt maps an EditorConfig number to setting id, ignoring a value the
// setting does not accept, as other values custos does not understand.
func setInt(v map[string]any, id string, n *int) {
	if d, _ := format.Lookup(id); n != nil && *n <= d.Max {
		v[id] = *n
	}
}

// FormatExcluded reports whether the format exclude globs of custos.json
// leave out path: a glob matches the path relative to the root, or one of
// its directories. Paths outside the root are never excluded.
func (c *Config) FormatExcluded(path string) bool {
	if len(c.Format.Exclude) == 0 {
		return false
	}
	abs, _ := filepath.Abs(path)
	rel, err := filepath.Rel(c.Root, abs)
	if err != nil || !insideRoot(rel) {
		return false
	}
	rel = filepath.ToSlash(rel)
	for _, glob := range c.Format.Exclude {
		for i := 1; i <= len(rel); i++ {
			if (i == len(rel) || rel[i] == '/') && editorconfig.Match(glob, rel[:i]) {
				return true
			}
		}
	}
	return false
}

// confinedRead reads .editorconfig files, those outside within (when set)
// counting as absent, symbolic links resolved.
func confinedRead(within string) func(string) ([]byte, error) {
	root := within
	if real, err := filepath.EvalSymlinks(within); err == nil {
		root = real
	}
	return func(name string) ([]byte, error) {
		if within != "" {
			real, err := filepath.EvalSymlinks(name)
			if err != nil {
				return nil, fs.ErrNotExist
			}
			if !inside(root, real) {
				return nil, fs.ErrNotExist
			}
			name = real
		}
		return safeio.ReadFile(name, MaxFileSize)
	}
}

// inside reports whether path lies in dir.
func inside(dir, path string) bool {
	rel, err := filepath.Rel(dir, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
