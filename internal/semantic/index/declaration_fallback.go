package index

import "strings"

// classAlias retains every alias declaration so removal and suppression can
// reveal an earlier declaration without mutating the extracted file symbols.
type classAlias struct {
	original, file, fallback string
}

// removeDeclarationCandidatesLocked groups repeated names so every candidate
// slice is filtered once. New slices preserve snapshots held by concurrent readers.
func (ix *Index) removeDeclarationCandidatesLocked(old *FileSymbols) {
	removed := make(map[*Constant]bool, len(old.Constants))
	for _, c := range old.Constants {
		removed[c] = true
	}
	constantKeys := map[string]bool{}
	for _, c := range old.Constants {
		k := strings.TrimPrefix(c.FQN, `\`)
		if constantKeys[k] {
			continue
		}
		constantKeys[k] = true
		cands := ix.constants[k]
		kept := make([]*Constant, 0, len(cands))
		for _, candidate := range cands {
			if !removed[candidate] {
				kept = append(kept, candidate)
			}
		}
		ix.constants[k] = kept
	}
	aliasKeys := map[string]bool{}
	for _, a := range old.ClassAliases {
		k := key(a[0])
		if aliasKeys[k] {
			continue
		}
		aliasKeys[k] = true
		cands := ix.aliases[k]
		kept := make([]classAlias, 0, len(cands))
		for _, candidate := range cands {
			if candidate.file != old.Path {
				kept = append(kept, candidate)
			}
		}
		ix.aliases[k] = kept
	}
}

// declarationShadowed checks the preferred function candidate of a namespace
// fallback call. As with Function's permissive availability policy, any source
// declaration counts. A buffer layer replacing a file hides its saved functions.
func (ix *Index) declarationShadowed(fqn string) bool {
	if fqn == "" {
		return false
	}
	k := key(fqn)
	for l := ix; l != nil; l = l.base {
		l.mu.RLock()
		found := false
		for _, f := range l.functions[k] {
			if !shadowedAbove(ix, l, f.File) {
				found = true
				break
			}
		}
		l.mu.RUnlock()
		if found {
			return true
		}
	}
	return false
}

// WithoutProvisionalDeclarations is a shallow view for single-file return
// annotation. A namespace-fallback define/class_alias can be shadowed by a
// function in another file, so its symbols cannot justify a cached contract.
// Functions and classes stay shared to receive the annotation's results.
func (fs *FileSymbols) WithoutProvisionalDeclarations() *FileSymbols {
	provisional := len(fs.ClassAliasFallbacks) > 0
	for _, c := range fs.Constants {
		provisional = provisional || c.DeclarationFallback != ""
	}
	if !provisional {
		return fs
	}
	view := *fs
	view.Constants = nil
	for _, c := range fs.Constants {
		if c.DeclarationFallback == "" {
			view.Constants = append(view.Constants, c)
		}
	}
	view.ClassAliases = nil
	view.ClassAliasFallbacks = nil
	for pos, a := range fs.ClassAliases {
		if fs.ClassAliasFallbacks[pos] == "" {
			view.ClassAliases = append(view.ClassAliases, a)
		}
	}
	return &view
}
