package flow

// bindings owns its map but shares immutable value cells with snapshots. A
// write replaces a shared cell; only cells still private to one frame change
// in place. Callers read value copies, never the cells themselves.
type bindingCell struct {
	Value
	owner *bool
}

type bindings map[string]*bindingCell

func (b bindings) lookup(name string) (Value, bool) {
	value := b[name]
	if value == nil {
		return Value{}, false
	}
	return value.Value, true
}

// Use a pointer merge for missing values so field-only reads can avoid a
// temporary full Value copy after inlining. This cell is never written.
var missingBinding Value

func (b bindings) cell(name string) *Value {
	value := b[name]
	if value == nil {
		return &missingBinding
	}
	return &value.Value
}

func (b bindings) value(name string) Value { return *b.cell(name) }
