package meta

import "testing"

func TestLookup(t *testing.T) {
	all, err := All()
	if err != nil || len(all) == 0 {
		t.Fatalf("All: %d, %v", len(all), err)
	}
	r, ok := Lookup(all[0].ID)
	if !ok || r.ID != all[0].ID {
		t.Fatal("lookup by ID")
	}
	if r2, ok := Lookup(all[0].LegacyID); !ok || r2 != r {
		t.Fatal("lookup by legacy ID")
	}
	if _, ok := Lookup("NoSuchRule"); ok {
		t.Fatal("unknown rule found")
	}
	if _, ok := Describe(all[0].ID); !ok {
		t.Fatal("description missing")
	}
}

func TestLoadBadCatalogue(t *testing.T) {
	saved, savedRules, savedByID, savedErr := rulesJSON, rules, byID, loadErr
	defer func() { rulesJSON, rules, byID, loadErr = saved, savedRules, savedByID, savedErr }()
	rulesJSON, loadErr = []byte("{"), nil
	load()
	if loadErr == nil {
		t.Fatal("corrupt catalogue must report an error")
	}
}
