package flow

import "testing"

func TestFrameValueCellsRemainIsolated(t *testing.T) {
	original := frame{}
	original.resetVars()
	initial := Value{Complete: true, Safe: HTML, Identity: 1}
	original.setVar("value", initial)
	initial.Identity = 99
	if original.vars.value("value").Identity != 1 {
		t.Fatal("write retained caller-owned storage")
	}
	original.setVar("zero", Value{})
	if _, known := original.vars.lookup("missing"); known {
		t.Fatal("missing variable became tracked")
	}
	if zero, known := original.vars.lookup("zero"); !known || zero.Complete {
		t.Fatal("tracked zero value became missing")
	}
	left := clone(&original)
	right := clone(&original)
	nested := clone(&left)
	original.setVar("value", Value{Complete: true, Identity: 2})
	left.unsetVar("value")
	right.setVar("other", Value{Complete: true, Identity: 3})
	if nested.vars.value("value").Identity != 1 || right.vars.value("value").Identity != 1 || original.vars.value("value").Identity != 2 {
		t.Fatal("write changed a retained snapshot")
	}
	if _, known := left.vars.lookup("value"); known {
		t.Fatal("unset retained the binding")
	}
	read := nested.vars.value("value")
	read.Safe = SQL
	read.Identity = 4
	if nested.vars.value("value").Safe != HTML || nested.vars.value("value").Identity != 1 {
		t.Fatal("reading exposed mutable storage")
	}
	nested.setVar("value", read)
	if right.vars.value("value").Identity != 1 || nested.vars.value("value").Identity != 4 {
		t.Fatal("nested write changed a sibling")
	}
	right.resetVars()
	if nested.vars.value("value").Identity != 4 || original.vars.value("zero").Complete {
		t.Fatal("reset changed retained values")
	}
	if _, known := nested.vars.lookup("other"); known {
		t.Fatal("new sibling binding leaked")
	}
}

func TestSharedFrameUnchangedWriteDoesNotAllocate(t *testing.T) {
	original := frame{}
	original.resetVars()
	value := Value{Complete: true, Safe: HTML}
	original.setVar("value", value)
	allocations := testing.AllocsPerRun(100, func() {
		branch := clone(&original)
		branch.setVar("value", value)
		if !branch.vars.value("value").Complete {
			t.Fatal("unchanged value lost")
		}
	})
	if allocations != 0 {
		t.Fatalf("unchanged write allocated %v objects", allocations)
	}
}

func TestOwnedFrameWriteDoesNotAllocate(t *testing.T) {
	original := frame{}
	original.resetVars()
	original.setVar("value", Value{Complete: true, Identity: 1})
	retained := original.vars.value("value")
	next := Value{Complete: true, Identity: 2}
	allocations := testing.AllocsPerRun(100, func() { original.setVar("value", next) })
	if allocations != 0 {
		t.Fatalf("owned write allocated %v objects", allocations)
	}
	if retained.Identity != 1 || original.vars.value("value").Identity != 2 {
		t.Fatal("owned write changed a read copy")
	}
}

func TestDetachedFrameUnchangedWriteDoesNotAllocate(t *testing.T) {
	original := frame{}
	original.resetVars()
	value := Value{Complete: true, Safe: HTML}
	original.setVar("value", value)
	branch := clone(&original)
	branch.setVar("other", Value{Complete: true})
	allocations := testing.AllocsPerRun(100, func() { branch.setVar("value", value) })
	if allocations != 0 {
		t.Fatalf("detached unchanged write allocated %v objects", allocations)
	}
	branch.setVar("value", Value{Complete: true, Safe: SQL})
	if original.vars.value("value").Safe != HTML {
		t.Fatal("detached write changed shared cell")
	}
}

func TestMissingFrameReadRemainsACopy(t *testing.T) {
	var vars bindings
	missing := vars.value("missing")
	missing.Safe = HTML
	if vars.value("missing").Safe != 0 {
		t.Fatal("missing read exposed global storage")
	}
}
