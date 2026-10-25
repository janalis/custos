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
	"time"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/platform/safeio"
	"custos/internal/project/config"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/stubs"
)

// Snapshot is immutable after construction. Its input manifest covers every
// configuration and semantic source read. Project inputs are retained in
// Sources; vendor files contribute declarations and a hash only.
type Snapshot struct {
	Engine  *analysis.Engine
	Index   *index.Index
	Sources map[string][]byte
	// Hashes is the manifest of retained inputs plus one VendorDigest entry.
	Hashes map[string]string
	// VendorHashes and Stats describe individual files for change detection.
	VendorHashes map[string]string
	Stats        map[string]FileStat
	Files        []string
	OwnFiles     []string
	Vendor       []string
	// Skipped lists the first vendor entries left out of semantic indexing
	// (symlinks, oversized or unparsable files); SkippedCount counts them all.
	Skipped      []string
	SkippedCount int
	Symbols      []*index.FileSymbols
	Results      []diagnostic.FileResult
	Config       *config.Config
	// Baseline is the captured baseline's key in Sources, "" without one.
	Baseline string
	// Paths are the requested inputs; empty selects the configured paths.
	Paths    []string
	Captured time.Time

	root          *os.Root
	cache         *VendorCache
	vendorSymbols map[string]*index.FileSymbols
}

// FileStat is the cheap identity of an input between two hash checks.
type FileStat struct {
	Size    int64
	ModTime int64
}

// VendorCache carries vendor hashes and declarations between snapshots of one
// root, so an unchanged dependency is neither read nor parsed again. It is not
// safe for concurrent snapshots.
type VendorCache struct {
	parse string
	files map[string]vendorEntry
}

type vendorEntry struct {
	stat    FileStat
	hash    string
	settled bool
	symbols *index.FileSymbols
}

const (
	MaxSnapshotBytes = 64 << 20
	MaxSnapshotFiles = 10000
	// VendorDigest is the Hashes key summarizing every vendor file hash.
	VendorDigest = "@vendor"
	maxSkipped   = 20
	// A file modified this close to its capture may change again without a
	// visible size or timestamp difference, so its stat is never trusted.
	racyWindow = 2 * time.Second
)

// Vendor files are read and hashed but not retained; tests lower the bounds.
var (
	maxVendorFiles       = 100000
	maxVendorBytes int64 = 1 << 30
)

var configInputs = []string{"custos.json", "composer.json", "composer.lock", "phpunit.xml", "phpunit.xml.dist"}

// ReadConfined never follows a link outside root or reads a special file.
func ReadConfined(root *os.Root, name string, limit int64) ([]byte, error) {
	return safeio.ReadRootFile(root, name, limit)
}

// Fingerprint identifies exact bytes; it does not store supplied buffers.
func Fingerprint(b []byte) string { sum := sha256.Sum256(b); return hex.EncodeToString(sum[:]) }

// BuildSnapshot constructs cancellable, root-confined project evidence.
func BuildSnapshot(ctx context.Context, root *os.Root, rules []analysis.Rule, paths []string, cache *VendorCache) (*Snapshot, error) {
	s, err := capture(ctx, root, rules, paths, cache, true)
	if err != nil {
		return nil, err
	}
	return AnalyzeSnapshot(ctx, s)
}

// CaptureSnapshot captures bounded inputs without parsing or analyzing PHP.
func CaptureSnapshot(ctx context.Context, root *os.Root, rules []analysis.Rule, paths []string, cache *VendorCache) (*Snapshot, error) {
	return capture(ctx, root, rules, paths, cache, false)
}

func capture(ctx context.Context, root *os.Root, rules []analysis.Rule, paths []string, cache *VendorCache, extractVendor bool) (*Snapshot, error) {
	cfg, err := config.LoadConfined(root)
	if err != nil {
		return nil, err
	}
	acfg := cfg.Analysis()
	acfg.ProjectPath = ProjectPaths(func(name string) bool {
		_, err := root.Stat(name)
		return err == nil
	}, ".")
	engine, err := analysis.NewEngine(rules, acfg)
	if err != nil {
		return nil, err
	}
	s := &Snapshot{Engine: engine, Index: index.New(stubs.Index()), Sources: map[string][]byte{}, Hashes: map[string]string{}, VendorHashes: map[string]string{}, Config: cfg, Paths: append([]string{}, paths...), Captured: time.Now(), root: root, cache: cache}
	total := 0
	read := func(name string, limit int64) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if _, ok := s.Sources[name]; ok {
			return nil
		}
		b, err := ReadConfined(root, name, limit)
		if err != nil {
			return err
		}
		total += len(b)
		if total > MaxSnapshotBytes || len(s.Sources) >= MaxSnapshotFiles {
			return fmt.Errorf("snapshot input budget exceeded: more than %d project files or %d MiB; analyze narrower paths or exclude directories", MaxSnapshotFiles, MaxSnapshotBytes>>20)
		}
		s.Sources[name] = b
		s.Hashes[name] = Fingerprint(b)
		return nil
	}
	for _, name := range configInputs {
		if err := read(name, config.MaxFileSize); err != nil {
			if !os.IsNotExist(err) {
				return nil, err
			}
			s.Hashes[name] = "absent"
		}
	}
	if cfg.Baseline != "" {
		// Settings may spell root-relative paths as "./x" or "x/".
		s.Baseline = path.Clean(filepath.ToSlash(cfg.Baseline))
		if err := read(s.Baseline, MaxSnapshotBytes); err != nil {
			return nil, err
		}
	}
	l, err := discover(ctx, root, cfg, engine.FilePatterns(), vendorDirBytes(s.Sources["composer.json"]), paths)
	if err != nil {
		return nil, err
	}
	for _, name := range l.own {
		if err := read(name, syntax.MaxFileSize); err != nil {
			return nil, err
		}
	}
	s.Files, s.OwnFiles, s.Vendor, s.Stats = l.selected, l.php, l.vendor, l.stats
	s.Skipped, s.SkippedCount = l.skipped, l.skippedCount
	if err := s.readVendor(ctx, extractVendor); err != nil {
		return nil, err
	}
	digest := sha256.New()
	for _, name := range s.Vendor {
		digest.Write([]byte(name + "\x00" + s.VendorHashes[name] + "\n"))
	}
	s.Hashes[VendorDigest] = hex.EncodeToString(digest.Sum(nil))
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

// listing is one deterministic walk of the project and vendor inputs.
type listing struct {
	own, selected, php, vendor []string
	stats                      map[string]FileStat
	skipped                    []string
	skippedCount               int
}

// found is what scanning one directory produced, in listing order.
type found struct {
	names []string
	infos []fs.FileInfo
	dirs  []string
	links []string
	err   error
}

// splitLevels is the depth at which a walk stops listing single directories
// and descends into whole subtrees.
const splitLevels = 2

func (l *listing) skip(name, reason string) {
	l.skippedCount++
	if len(l.skipped) < maxSkipped {
		l.skipped = append(l.skipped, name+" ("+reason+")")
	}
}

// discover lists inputs without reading them. Symlinks are never followed:
// one under a project path is an error, one under vendor is skipped.
func discover(ctx context.Context, root *os.Root, cfg *config.Config, patterns []string, vendor string, paths []string) (*listing, error) {
	if len(paths) == 0 {
		paths = cfg.Paths
	}
	l := &listing{stats: map[string]FileStat{}}
	packageToSkip := path.Join(vendor, "janalis/custos")
	own, selected, php := map[string]bool{}, map[string]bool{}, map[string]bool{}
	vendorBytes := int64(0)
	walk := func(base string, symbolsOnly bool) error {
		base = path.Clean(filepath.ToSlash(base))
		if base != "." && !fs.ValidPath(base) {
			return fmt.Errorf("invalid relative path")
		}
		inVendor := symbolsOnly && base == vendor
		scanned := base
		file := func(name string, info fs.FileInfo) {
			match := strings.HasSuffix(name, ".php")
			if !symbolsOnly {
				match = match || matchesAny(path.Base(name), patterns)
			}
			if !match {
				return
			}
			if inVendor {
				if info.Size() > syntax.MaxFileSize {
					l.skip(name, "too large")
					return
				}
				vendorBytes += info.Size()
				l.vendor = append(l.vendor, name)
			} else {
				own[name] = true
				if strings.HasSuffix(name, ".php") {
					php[name] = true
				}
				if !symbolsOnly {
					selected[name] = true
				}
			}
			l.stats[name] = FileStat{Size: info.Size(), ModTime: info.ModTime().UnixNano()}
		}
		// Each directory is opened once and its entries are reached by a
		// single path component, which keeps confined traversal cheap.
		scan := func(base string, recurse bool) *found {
			f := &found{}
			var visit func(parent *os.Root, entry, name string) error
			visit = func(parent *os.Root, entry, name string) error {
				dir, err := parent.OpenRoot(entry)
				if err != nil {
					return err
				}
				defer dir.Close()
				entries, err := fs.ReadDir(dir.FS(), ".")
				for _, d := range entries {
					child := path.Join(name, d.Name())
					if child == packageToSkip || child == vendor {
						continue
					}
					if d.Type()&os.ModeSymlink != 0 {
						if inVendor {
							f.links = append(f.links, child)
							continue
						}
						return fmt.Errorf("symlink input is not supported: %s", child)
					}
					if !d.IsDir() {
						// A file removed since the listing is simply not an input.
						if info, err := d.Info(); err == nil {
							f.names, f.infos = append(f.names, child), append(f.infos, info)
						}
						continue
					}
					if inVendor && (d.Name() == "tests" || d.Name() == "Tests" || d.Name() == "test") || excluded(cfg.Exclude, d.Name(), strings.TrimPrefix(child, scanned+"/")) || hidden(d.Name()) {
						continue
					}
					if !recurse {
						f.dirs = append(f.dirs, child)
						continue
					}
					if err := visit(dir, d.Name(), child); err != nil {
						return err
					}
				}
				return err
			}
			f.err = visit(root, base, base)
			return f
		}
		// The top levels are listed one directory at a time so that the
		// subtrees below them can be walked in parallel; results are merged
		// in listing order.
		tree := func() error {
			dirs := []string{base}
			for level := 0; len(dirs) > 0; level++ {
				out := make([]*found, len(dirs))
				if err := Each(ctx, len(dirs), func(i int) { out[i] = scan(dirs[i], level == splitLevels) }); err != nil {
					return err
				}
				dirs = nil
				for _, f := range out {
					if f.err != nil {
						return f.err
					}
					for i, name := range f.names {
						file(name, f.infos[i])
					}
					for _, name := range f.links {
						l.skip(name, "symlink")
					}
					dirs = append(dirs, f.dirs...)
				}
			}
			return nil
		}
		info, err := root.Lstat(base)
		switch {
		case os.IsNotExist(err):
			return fmt.Errorf("path not found: %s", base)
		case err != nil:
			return err
		case info.Mode()&os.ModeSymlink != 0:
			return fmt.Errorf("symlink input is not supported: %s", base)
		case !info.IsDir():
			file(base, info)
			return nil
		}
		return tree()
	}
	for _, p := range paths {
		if err := walk(p, false); err != nil {
			return nil, err
		}
	}
	for _, p := range cfg.Paths {
		if err := walk(p, true); err != nil {
			return nil, err
		}
	}
	if _, err := root.Stat(vendor); err == nil {
		if err := walk(vendor, true); err != nil {
			return nil, err
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	if len(l.vendor) > maxVendorFiles || vendorBytes > maxVendorBytes {
		return nil, fmt.Errorf("vendor input budget exceeded: %d files, %d MiB (limits %d files, %d MiB)", len(l.vendor), vendorBytes>>20, maxVendorFiles, maxVendorBytes>>20)
	}
	l.own, l.selected, l.php = sorted(own), sorted(selected), sorted(php)
	sort.Strings(l.vendor)
	return l, nil
}

func excluded(patterns []string, entry, name string) bool {
	for _, ex := range patterns {
		if entry == ex || name == strings.TrimSuffix(ex, "/") {
			return true
		}
	}
	return false
}

func sorted(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for name := range set {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func (s *Snapshot) parseOptions() syntax.Options {
	return syntax.Options{Version: s.Config.PHP, ShortOpenTag: s.Config.ShortOpenTag}
}

func (s *Snapshot) inferOptions() infer.Options {
	return infer.Options{Assertions: s.Config.Flow.Assertions}
}

// readVendor hashes every vendor file, verifying hashes already in the
// manifest, and extracts declarations when asked. Cached entries whose stat
// is unchanged are trusted without a read.
func (s *Snapshot) readVendor(ctx context.Context, extractSymbols bool) error {
	opt := s.parseOptions()
	// Declarations depend on the inference options too (inferred returns).
	parse := fmt.Sprint(opt.Version, opt.ShortOpenTag, s.inferOptions())
	var cached map[string]vendorEntry
	if s.cache != nil && s.cache.parse == parse {
		cached = s.cache.files
	}
	entries := make([]vendorEntry, len(s.Vendor))
	errs := make([]error, len(s.Vendor))
	// Vendor is sorted, so each directory is one run: open it once and read
	// its files by base name.
	var runs [][2]int
	for i, name := range s.Vendor {
		if i == 0 || path.Dir(name) != path.Dir(s.Vendor[i-1]) {
			runs = append(runs, [2]int{i, i})
		}
		runs[len(runs)-1][1] = i + 1
	}
	err := Each(ctx, len(runs), func(r int) {
		var dir *os.Root
		defer func() {
			if dir != nil {
				_ = dir.Close()
			}
		}()
		for i := runs[r][0]; i < runs[r][1]; i++ {
			name := s.Vendor[i]
			st := s.Stats[name]
			old, known := cached[name]
			if known && old.settled && old.stat == st && (!extractSymbols || old.symbols != nil) {
				entries[i] = old
				continue
			}
			if dir == nil {
				if dir, errs[i] = s.root.OpenRoot(path.Dir(name)); errs[i] != nil {
					return
				}
			}
			b, err := ReadConfined(dir, path.Base(name), syntax.MaxFileSize)
			if err != nil {
				errs[i] = err
				return
			}
			e := vendorEntry{stat: st, hash: Fingerprint(b), settled: time.Unix(0, st.ModTime).Before(s.Captured.Add(-racyWindow))}
			if known && old.hash == e.hash {
				e.symbols = old.symbols
			}
			if extractSymbols && e.symbols == nil {
				e.symbols, _ = extractCaptured(name, b, opt, s.inferOptions())
			}
			entries[i] = e
		}
	})
	if err != nil {
		return err
	}
	symbols := map[string]*index.FileSymbols{}
	files := make(map[string]vendorEntry, len(s.Vendor))
	for i, name := range s.Vendor {
		if errs[i] != nil {
			return errs[i]
		}
		e := entries[i]
		if known, ok := s.VendorHashes[name]; ok && known != e.hash {
			return fmt.Errorf("input changed during analysis: %s", name)
		}
		s.VendorHashes[name] = e.hash
		if extractSymbols {
			if e.symbols == nil {
				s.SkippedCount++
				if len(s.Skipped) < maxSkipped {
					s.Skipped = append(s.Skipped, name+" (unparsable)")
				}
			} else {
				symbols[name] = e.symbols
			}
		}
		// DropStaleInferred rewrites declarations with return dependencies,
		// so only dependency-free ones are shared between snapshots.
		if e.symbols != nil && len(e.symbols.ReturnDeps) > 0 {
			e.symbols = nil
		}
		files[name] = e
	}
	if extractSymbols {
		s.vendorSymbols = symbols
	}
	if s.cache != nil {
		s.cache.parse, s.cache.files = parse, files
	}
	return nil
}

// Source returns the captured bytes of a project input, or re-reads a vendor
// file and checks it against the manifest.
func (s *Snapshot) Source(name string) ([]byte, error) {
	if b, ok := s.Sources[name]; ok {
		return b, nil
	}
	hash, ok := s.VendorHashes[name]
	if !ok {
		return nil, fmt.Errorf("path unavailable: %s", name)
	}
	b, err := ReadConfined(s.root, name, syntax.MaxFileSize)
	if err != nil {
		return nil, err
	}
	if Fingerprint(b) != hash {
		return nil, fmt.Errorf("input changed: %s", name)
	}
	return b, nil
}

// ChangedInputs reports inputs that differ from the snapshot manifest. A file
// whose size and modification time are unchanged is trusted unless it was
// modified within racyWindow of the capture; every other file is hashed.
func ChangedInputs(ctx context.Context, s *Snapshot) ([]string, error) {
	var changed []string
	for _, name := range append(append([]string{}, configInputs...), s.Baseline) {
		if name == "" {
			continue
		}
		hash := "absent"
		b, err := ReadConfined(s.root, name, MaxSnapshotBytes)
		if err == nil {
			hash = Fingerprint(b)
		} else if !os.IsNotExist(err) {
			return nil, err
		}
		if hash != s.Hashes[name] {
			changed = append(changed, name)
		}
	}
	if len(changed) > 0 {
		return changed, nil
	}
	l, err := discover(ctx, s.root, s.Config, s.Engine.FilePatterns(), vendorDirBytes(s.Sources["composer.json"]), s.Paths)
	if err != nil {
		return nil, err
	}
	settled := s.Captured.Add(-racyWindow)
	seen := map[string]bool{}
	check := func(names []string, hashes map[string]string) error {
		for _, name := range names {
			seen[name] = true
			hash, known := hashes[name]
			st := l.stats[name]
			if known && st == s.Stats[name] && time.Unix(0, st.ModTime).Before(settled) {
				continue
			}
			if known {
				b, err := ReadConfined(s.root, name, syntax.MaxFileSize)
				if err != nil {
					return err
				}
				known = Fingerprint(b) == hash
			}
			if !known {
				changed = append(changed, name)
			}
		}
		return nil
	}
	if err := check(l.own, s.Hashes); err != nil {
		return nil, err
	}
	if err := check(l.vendor, s.VendorHashes); err != nil {
		return nil, err
	}
	for name := range s.Stats {
		if !seen[name] {
			changed = append(changed, name)
		}
	}
	sort.Strings(changed)
	return changed, nil
}

// AnalyzeSnapshot publishes index and flow copies before producing findings.
func AnalyzeSnapshot(ctx context.Context, s *Snapshot) (*Snapshot, error) {
	if s.vendorSymbols == nil && len(s.Vendor) > 0 {
		if err := s.readVendor(ctx, true); err != nil {
			return nil, err
		}
	}
	opt := s.parseOptions()
	var semantic []string
	for name := range s.Sources {
		if strings.HasSuffix(name, ".php") {
			semantic = append(semantic, name)
		}
	}
	sort.Strings(semantic)
	extracted := make([]*index.FileSymbols, len(semantic))
	errs := make([]error, len(semantic))
	if err := Each(ctx, len(semantic), func(i int) {
		extracted[i], errs[i] = extractCaptured(semantic[i], s.Sources[semantic[i]], opt, s.inferOptions())
	}); err != nil {
		return nil, err
	}
	all := map[string]*index.FileSymbols{}
	for i, name := range semantic {
		if errs[i] != nil {
			return nil, errs[i]
		}
		all[name] = extracted[i]
	}
	names := semantic
	for name, symbols := range s.vendorSymbols {
		if all[name] == nil {
			all[name] = symbols
			names = append(names, name)
		}
	}
	// Declarations are added in path order so duplicates resolve identically.
	sort.Strings(names)
	s.Index.DeferTypeAliases()
	for _, name := range names {
		s.Symbols = append(s.Symbols, all[name])
		s.Index.Add(all[name])
	}
	for _, symbols := range s.Symbols {
		s.Index.DropStaleInferred(symbols)
	}
	s.Index.LinkTypeAliases()
	s.Index.Freeze()
	s.Engine = s.Engine.WithIndex(s.Index)

	if s.Engine.NeedsFlow() {
		var srcs [][]byte
		for _, name := range s.OwnFiles {
			srcs = append(srcs, s.Sources[name])
		}
		s.Engine = s.Engine.WithFlow(BuildFlowSnapshot(ctx, s.OwnFiles, srcs, opt, s.Index, s.Config.Flow))
	}
	results := make([]diagnostic.FileResult, len(s.Files))
	if err := Each(ctx, len(s.Files), func(i int) {
		results[i] = AnalyzeBuffer(s.Engine, s.Files[i], s.Sources[s.Files[i]], opt)
		dropEdits(results[i].Findings)
	}); err != nil {
		return nil, err
	}
	s.Results = results
	return s, nil
}

// Captured sources receive the same index crash isolation as disk sources.
func extractCaptured(name string, src []byte, opt syntax.Options, options infer.Options) (symbols *index.FileSymbols, err error) {
	defer func() {
		if recover() != nil {
			symbols = nil
			err = fmt.Errorf("semantic extraction failed for %s", name)
		}
	}()
	return extract(name, src, opt, options), nil
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
