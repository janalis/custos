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
	for _, fact := range all {
		if fact.LegacyID != "" {
			if r2, ok := Lookup(fact.LegacyID); !ok || r2.ID != fact.ID {
				t.Fatal("lookup by legacy ID")
			}
		}
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

func TestDecodeCatalogues(t *testing.T) {
	all, lookup, err := decodeCatalogues([]byte(`[{"id":"Zulu","legacyId":"ZuluInspection"}]`), []byte(`[{"id":"Alpha"}]`))
	if err != nil || len(all) != 2 || all[0].ID != "Alpha" || !all[0].Native || all[1].Native {
		t.Fatalf("merged catalogue: %+v, %v", all, err)
	}
	if lookup["ZuluInspection"] != lookup["Zulu"] || lookup["Alpha"] != &all[0] {
		t.Fatal("aliases or native lookup lost")
	}
	if _, ok := lookup[""]; ok {
		t.Fatal("empty legacy ID registered")
	}
	for _, tc := range []struct{ extracted, native string }{
		{`{`, `[]`},
		{`[]`, `{`},
		{`[{"id":""}]`, `[]`},
		{`[{"id":"Alpha"}]`, `[{"id":"Alpha"}]`},
		{`[{"id":"Alpha","legacyId":"Beta"}]`, `[{"id":"Beta"}]`},
	} {
		if _, _, err := decodeCatalogues([]byte(tc.extracted), []byte(tc.native)); err == nil {
			t.Errorf("invalid catalogues accepted: %+v", tc)
		}
	}
}
