package flow

import (
	"reflect"
	"strings"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
)

// Sink describes an external operation reached inside a project wrapper.
type Sink struct {
	Name     string
	Context  Context
	Argument Value
	Path     string
	Span     syntax.Span
}

// Summary is a compact, syntax-free callable return and sink contract.
type Summary struct {
	Return    Value
	Sinks     []Sink
	Generator bool
}

// File contains summaries from one indexed file. It retains no source or AST.
type File struct {
	Path      string
	Summaries map[string]Summary
}

// Snapshot is an immutable project-local summary collection. WithFile and
// WithoutFile return copies; readers never observe mutations or need locks.
type Snapshot struct {
	files     map[string]*File
	summaries map[string]Summary
}

// NewSnapshot creates an empty summary collection.
func NewSnapshot(files ...*File) *Snapshot {
	s := &Snapshot{files: map[string]*File{}, summaries: map[string]Summary{}}
	for _, f := range files {
		if f != nil {
			s.files[f.Path] = f
		}
	}
	s.rebuild()
	return s
}

// WithFile replaces all summaries of a file in a new snapshot.
func (s *Snapshot) WithFile(f *File) *Snapshot {
	n := NewSnapshot()
	if s != nil {
		for path, file := range s.files {
			if path != f.Path {
				n.files[path] = file
			}
		}
	}
	n.files[f.Path] = f
	n.rebuild()
	return n
}

// WithoutFile removes all summaries of a file in a new snapshot.
func (s *Snapshot) WithoutFile(path string) *Snapshot {
	n := NewSnapshot()
	if s != nil {
		for p, f := range s.files {
			if p != path {
				n.files[p] = f
			}
		}
	}
	n.rebuild()
	return n
}

func (s *Snapshot) rebuild() {
	duplicates := map[string]bool{}
	for _, f := range s.files {
		for name, summary := range f.Summaries {
			if _, exists := s.summaries[name]; exists {
				duplicates[name] = true
			}
			s.summaries[name] = summary
		}
	}
	for name := range duplicates {
		delete(s.summaries, name)
	}
}

func (s *Snapshot) lookup(name string) (Summary, bool) {
	if s == nil {
		return Summary{}, false
	}
	v, ok := s.summaries[strings.ToLower(name)]
	return v, ok
}

// Equal reports whether two immutable snapshots expose the same contracts.
func (s *Snapshot) Equal(other *Snapshot) bool {
	if s == nil || other == nil {
		return s == other
	}
	return reflect.DeepEqual(s.summaries, other.summaries)
}

// Extract computes compact summaries while the project's parser tree is
// available. It does not open any other files. A supplied snapshot permits
// bounded composition with previously summarized project wrappers.
func Extract(file *syntax.File, ix *index.Index, php phpversion.Version, snapshot *Snapshot) *File {
	types := infer.NewEnv(file, names.New(file), ix, php)
	e := New(file, types, snapshot)
	out := &File{Path: file.Path, Summaries: map[string]Summary{}}
	syntax.InspectFile(file, func(n syntax.Node) bool {
		name := ""
		switch n := n.(type) {
		case *syntax.Function:
			if n.Name != nil {
				name, _ = types.Names.Function(n.Name.Value, n.Span().Start)
			}
		case *syntax.Method:
			if n.Name != nil {
				name = types.ClassFQN(syntax.EnclosingClass(n)) + "::" + n.Name.Value
			}
		}
		if name != "" {
			r := e.scope(n)
			sum := Summary{Return: Value{Complete: r.complete}}
			body := syntax.VariableScopeBody(n)
			if body != nil {
				syntax.Inspect(body, func(child syntax.Node) bool {
					if child != body && syntax.IsVariableScope(child) {
						return false
					}
					switch child.(type) {
					case *syntax.Yield, *syntax.YieldFrom:
						sum.Generator = true
					}
					return true
				})
			}
			for i, v := range r.returns {
				if i == 0 {
					sum.Return = v
				} else {
					sum.Return = merge(sum.Return, v)
				}
			}
			sum.Return.Complete = sum.Return.Complete && r.complete
			sum.Return.Expr = nil
			sum.Return.Identity = 0
			if !sum.Generator {
				for _, c := range r.calls {
					sum.Sinks = append(sum.Sinks, e.nativeSinks(c)...)
					sum.Sinks = append(sum.Sinks, e.Sinks(c.Node)...)
				}
			}
			out.Summaries[strings.ToLower(name)] = sum
		}
		return true
	})
	return out
}

// nativeSinks recognizes resolved native sinks and binds named arguments.
func (e *Env) nativeSinks(c Call) []Sink {
	name, builtin := e.callName(c.Node)
	if !builtin {
		return nil
	}
	args, known := e.bind(c.Node, c.Arguments)
	if !known || len(args) == 0 {
		return nil
	}
	context := Context(0)
	position := 0
	switch c.Node.(type) {
	case *syntax.MethodCall:
		_, qualified, resolved := e.declaration(c.Node)
		if resolved && strings.HasPrefix(qualified, "pdo::") && (name == "query" || name == "exec" || name == "prepare") {
			context = SQL
		}
	case *syntax.FuncCall:
		switch name {
		case "system", "exec", "shell_exec", "passthru", "popen":
			context = Shell
		case "header":
			context = Header
		case "readfile", "unlink", "file_get_contents", "fopen", "file_put_contents", "rmdir":
			context = Path
		case "curl_init":
			context = URL
		case "curl_setopt":
			if len(args) > 2 {
				if constant, ok := args[1].Expr.(*syntax.ConstFetch); ok && strings.EqualFold(constant.Name.Value, "CURLOPT_URL") {
					context = URL
					position = 2
				}
			}
		}
	}
	if context == 0 {
		return nil
	}
	value := args[position]
	value.Expr = nil
	value.Identity = 0
	sinks := []Sink{{Name: name, Context: context, Argument: value, Path: e.file.Path, Span: c.Node.Span()}}
	if name == "header" {
		if binary, ok := args[0].Expr.(*syntax.Binary); ok && binary.Op.Kind == syntax.TDot {
			if prefix, known := literalText(binary.Left); known && strings.HasPrefix(strings.ToLower(prefix), "location:") {
				sinks = append(sinks, Sink{Name: name, Context: URL, Argument: value, Path: e.file.Path, Span: c.Node.Span()})
			}
		}
	}
	return sinks
}

// Sinks expands known project wrapper sink contracts at this call site.
// The argument sources and escaping contexts come from this call's flow.
func (e *Env) Sinks(at syntax.Node) []Sink {
	name := e.summaryName(at)
	summary, ok := e.snapshot.lookup(name)
	if !ok {
		return nil
	}
	args, known := e.BoundArguments(at)
	if !known {
		return nil
	}
	out := make([]Sink, 0, len(summary.Sinks))
	for _, s := range summary.Sinks {
		v := Value{Complete: s.Argument.Complete}
		for _, source := range s.Argument.Sources {
			if source.Parameter >= 0 {
				if source.Parameter < len(args) {
					v = merge(v, args[source.Parameter])
				} else {
					v.Complete = false
				}
			} else {
				v.Sources = append(v.Sources, source)
			}
		}
		v.Safe |= s.Argument.Safe
		s.Argument = v
		out = append(out, s)
	}
	return out
}
