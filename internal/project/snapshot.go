package project

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/platform/safeio"
	"custos/internal/project/config"
	"custos/internal/semantic/index"
	"custos/internal/semantic/stubs"
)

// Snapshot is immutable after construction. Its input manifest includes every
// configuration and semantic source read, including vendor declarations.
type Snapshot struct {
	Engine   *analysis.Engine
	Index    *index.Index
	Sources  map[string][]byte
	Hashes   map[string]string
	Files    []string
	OwnFiles []string
	Symbols  []*index.FileSymbols
	Results  []diagnostic.FileResult
	Config   *config.Config
	// Baseline is the captured baseline's key in Sources, "" without one.
	Baseline string
}

const (
	MaxSnapshotBytes = 64 << 20
	MaxSnapshotFiles = 10000
)

// ReadConfined never follows a link outside root or reads a special file.
func ReadConfined(root *os.Root, name string, limit int64) ([]byte, error) {
	return safeio.ReadRootFile(root, name, limit)
}

// Fingerprint identifies exact bytes; it does not store supplied buffers.
func Fingerprint(b []byte) string { sum := sha256.Sum256(b); return hex.EncodeToString(sum[:]) }

// BuildSnapshot constructs cancellable, root-confined project evidence.
func BuildSnapshot(ctx context.Context, root *os.Root, rules []analysis.Rule, paths []string) (*Snapshot, error) {
	s, err := CaptureSnapshot(ctx, root, rules, paths)
	if err != nil {
		return nil, err
	}
	return AnalyzeSnapshot(ctx, s)
}

// CaptureSnapshot captures bounded inputs without parsing or analyzing PHP.
func CaptureSnapshot(ctx context.Context, root *os.Root, rules []analysis.Rule, paths []string) (*Snapshot, error) {
	cfg, err := config.LoadConfined(root)
	if err != nil {
		return nil, err
	}
	engine, err := analysis.NewEngine(rules, cfg.Analysis())
	if err != nil {
		return nil, err
	}
	s := &Snapshot{Engine: engine, Index: index.New(stubs.Index()), Sources: map[string][]byte{}, Hashes: map[string]string{}, Config: cfg}
	total := 0
	read := func(name string, limit int64) ([]byte, error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if b, ok := s.Sources[name]; ok {
			return b, nil
		}
		b, err := ReadConfined(root, name, limit)
		if err != nil {
			return nil, err
		}
		total += len(b)
		if total > MaxSnapshotBytes || len(s.Sources) >= MaxSnapshotFiles {
			return nil, fmt.Errorf("snapshot input budget exceeded")
		}
		s.Sources[name] = b
		s.Hashes[name] = Fingerprint(b)
		return b, nil
	}
	for _, name := range []string{"custos.json", "composer.json", "composer.lock"} {
		if _, err := read(name, config.MaxFileSize); err != nil {
			if !os.IsNotExist(err) {
				return nil, err
			}
			s.Hashes[name] = "absent"
		}
	}
	if cfg.Baseline != "" {
		// Settings may spell root-relative paths as "./x" or "x/".
		s.Baseline = path.Clean(filepath.ToSlash(cfg.Baseline))
		if _, err := read(s.Baseline, MaxSnapshotBytes); err != nil {
			return nil, err
		}
	}
	if len(paths) == 0 {
		paths = cfg.Paths
	}
	vendor := "vendor"
	if b := s.Sources["composer.json"]; b != nil {
		vendor = vendorDirBytes(b)
	}
	packageToSkip := path.Join(vendor, "janalis/custos")
	selected := map[string]bool{}
	own := map[string]bool{}
	discover := func(base string, symbolsOnly bool) error {
		base = path.Clean(filepath.ToSlash(base))
		if base != "." && !fs.ValidPath(base) {
			return fmt.Errorf("invalid relative path")
		}
		return fs.WalkDir(root.FS(), base, func(name string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if err := ctx.Err(); err != nil {
				return err
			}
			if name == packageToSkip || (name == vendor && name != base) {
				if d.IsDir() {
					return fs.SkipDir
				}
				return nil
			}
			if d.Type()&os.ModeSymlink != 0 {
				return fmt.Errorf("symlink input is not supported: %s", name)
			}
			if d.IsDir() {
				if name != base {
					if symbolsOnly && base == vendor && (d.Name() == "tests" || d.Name() == "Tests" || d.Name() == "test") {
						return fs.SkipDir
					}
					for _, ex := range cfg.Exclude {
						if d.Name() == ex || name == strings.TrimSuffix(ex, "/") {
							return fs.SkipDir
						}
					}
				}
				return nil
			}
			match := strings.HasSuffix(name, ".php")
			if !symbolsOnly {
				for _, pat := range engine.FilePatterns() {
					ok, _ := path.Match(pat, d.Name())
					match = match || ok
				}
			}
			if match {
				if _, err := read(name, syntax.MaxFileSize); err != nil {
					return err
				}
				if strings.HasSuffix(name, ".php") && base != vendor {
					own[name] = true
				}
				if !symbolsOnly {
					selected[name] = true
				}
			}
			return nil
		})
	}
	for _, p := range paths {
		if err := discover(p, false); err != nil {
			return nil, err
		}
	}
	for _, p := range cfg.Paths {
		if err := discover(p, true); err != nil {
			return nil, err
		}
	}
	if _, err := root.Stat(vendor); err == nil {
		if err := discover(vendor, true); err != nil {
			return nil, err
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	for name := range selected {
		s.Files = append(s.Files, name)
	}
	sort.Strings(s.Files)
	for name := range own {
		s.OwnFiles = append(s.OwnFiles, name)
	}
	sort.Strings(s.OwnFiles)
	// Resolve from the captured bytes to avoid configuration races.
	captured, err := config.ResolveCaptured(root.Name(), s.Sources)
	if err != nil {
		return nil, err
	}
	a, _ := json.Marshal(cfg)
	b, _ := json.Marshal(captured)
	if string(a) != string(b) {
		return nil, fmt.Errorf("configuration changed during snapshot capture")
	}
	return s, nil
}

// AnalyzeSnapshot publishes index and flow copies before producing findings.
func AnalyzeSnapshot(ctx context.Context, s *Snapshot) (*Snapshot, error) {
	cfg := s.Config
	var semantic []string
	for name := range s.Sources {
		if strings.HasSuffix(name, ".php") {
			semantic = append(semantic, name)
		}
	}
	sort.Strings(semantic)
	opt := syntax.Options{Version: cfg.PHP, ShortOpenTag: cfg.ShortOpenTag}
	for _, name := range semantic {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		symbols, err := extractCaptured(name, s.Sources[name], opt)
		if err != nil {
			return nil, err
		}
		s.Symbols = append(s.Symbols, symbols)
		s.Index.Add(symbols)
	}
	for _, symbols := range s.Symbols {
		s.Index.DropStaleInferred(symbols)
	}
	s.Engine = s.Engine.WithIndex(s.Index)

	if s.Engine.NeedsFlow() {
		var srcs [][]byte
		for _, name := range s.OwnFiles {
			srcs = append(srcs, s.Sources[name])
		}
		s.Engine = s.Engine.WithFlow(BuildFlowSnapshot(s.OwnFiles, srcs, opt, s.Index))
	}
	for _, name := range s.Files {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		result := AnalyzeBuffer(s.Engine, name, s.Sources[name], opt)
		dropEdits(result.Findings)
		s.Results = append(s.Results, result)
	}
	return s, nil
}

// Captured sources receive the same index crash isolation as disk sources.
func extractCaptured(name string, src []byte, opt syntax.Options) (symbols *index.FileSymbols, err error) {
	defer func() {
		if recover() != nil {
			symbols = nil
			err = fmt.Errorf("semantic extraction failed for %s", name)
		}
	}()
	return extract(name, src, opt), nil
}

// OverlaySnapshot replaces one file with supplied bytes and rebuilds semantic
// publication using captured sources only. Disk hashes remain the input manifest.
func OverlaySnapshot(ctx context.Context, s *Snapshot, name string, src []byte) (*Snapshot, error) {
	total := len(src)
	for path, b := range s.Sources {
		if path != name {
			total += len(b)
		}
	}
	if total > MaxSnapshotBytes || (s.Sources[name] == nil && len(s.Sources) >= MaxSnapshotFiles) {
		return nil, fmt.Errorf("buffer snapshot budget exceeded")
	}
	s.Sources[name] = src
	found := false
	for _, path := range s.OwnFiles {
		if path == name {
			found = true
		}
	}
	if !found {
		s.OwnFiles = append(s.OwnFiles, name)
		sort.Strings(s.OwnFiles)
	}
	s.Files = []string{name}
	s.Results = nil
	s.Symbols = nil
	s.Index = index.New(stubs.Index())
	return AnalyzeSnapshot(ctx, s)
}
