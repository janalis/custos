package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"custos/internal/diagnostic"
	"custos/internal/inspection/meta"
)

// Everything below is an invented plugin layout with made-up rule names.

const pluginBody = `<idea-plugin>
  <name>Fake Garden Plugin</name>
  <extensions defaultExtensionNs="com.example">
    <localInspection language="PHP" shortName="AlphaInspection" groupName="Garden"
        level="WARNING" enabledByDefault="true" implementationClass="org.fake.garden.AlphaInspection"/>
    <localInspection language="PHP" shortName="BetaInspection" groupName="Shed"
        level="ERROR" enabledByDefault="false" implementationClass="org.fake.BetaInspection"/>
    <localInspection language="PHP" shortName="GammaInspection" groupName="Garden"
        level="WEAK WARNING" enabledByDefault="true" implementationClass="org.fake.garden.GammaInspection"/>
    <!--
    <localInspection language="PHP" shortName="DeltaInspection" groupName="Shed"
        level="INFO" implementationClass="org.fake.DeltaInspection"/>
    <localInspection language="PHP" shortName="Bad&Inspection" implementationClass="org.fake.X"/>
    <localInspection language="PHP" groupName="Shed" implementationClass="org.fake.Y"/>
    -->
  </extensions>
</idea-plugin>
`

const rulesBody = `# Fake rules

| Group | Short name | Notes | QF | Extra |
| ----- | ---------- | ----- | -- | ----- |
| Garden | AlphaInspection | waters plants | yes | - |
| Shed | BetaInspection | stacks pots | no | - |
| too | short |
`

const alphaSrc = `package org.fake.garden;
public class AlphaInspection {
    public boolean LOUD = true;
    public int LIMIT = 5;
    public int ODD = SOME_CONSTANT;
    public String LABEL = "sprout";
    public final PhpUnitVersion FLAVOUR = PhpUnitVersion.PHPUNIT42;
    public List<String> WORDS = new ArrayList<>();
    private boolean hidden = false;
}
`

// gardenTest exercises every statement form parseMethod understands.
const gardenTest = `package org.fake.garden;
// a stray comment mentioning public void testCommented()
/* block comment: public void testAlsoCommented() { } */
public class GardenTest {
    private AlphaInspection quietAlpha() {
        final AlphaInspection a = new AlphaInspection();
        a.LOUD = false;
        a.addWord("moss");
        b.LIMIT = 3;
        b.addWord("fern");
        return a;
    }
    private Shovel notARule() {
        Shovel s = new Shovel();
        return s;
    }
    public void testBasics() {
        PhpLanguageLevel level = PhpLanguageLevel.parse("7.4");
        myFixture.setLanguageLevel(level);
        AlphaInspection alpha = new AlphaInspection();
        Shovel shovel = new Shovel();
        alpha.LIMIT = 9;
        other.LIMIT = 1;
        alpha.addWord("ivy");
        other.addWord("x");
        ComparisonStyle.force(ComparisonStyle.YODA);
        myFixture.enableInspections(alpha, new BetaInspection(), new Rake(), quietAlpha(), quietAlpha, nothing(), stray);
        myFixture.configureByFile("garden/basic.php");
        myFixture.testHighlighting(true, false, true);
        myFixture.getAllQuickFixes().forEach(fix -> myFixture.launchAction(fix));
        myFixture.setTestDataPath(".");
        myFixture.checkResultByFile("garden/basic.fixed.php");
    }
    public void testLevels() {
        myFixture.setLanguageLevel(PhpLanguageLevel.PHP560);
        myFixture.setLanguageLevel(unknownLevel);
        myFixture.setLanguageLevel(levelOf(7));
        myFixture.enableInspections(new GammaInspection());
        myFixture.configureByFile("garden/first.php");
        PhpLanguageLevel.set(PhpLanguageLevel.PHP710);
        myFixture.configureByFile("garden/second.php");
        myFixture.testHighlighting(true, false, true);
        PhpLanguageLevel.set(null);
        if (enabled) { myFixture.configureByFile("garden/third.php");
        Object o = new Object() { int unused };
        myFixture.checkResultByFile("garden/third.php", "garden/third.fixed.php", true);
    }
    public void testNothing() {
        myFixture.enableInspections();
        myFixture.testHighlighting(true, false, true);
        myFixture.checkResultByFile("garden/none.fixed.php");
    }
}
`

const shedTest = `package org.fake;
public class ShedTest {
    public void testPots() {
        myFixture.enableInspections(new BetaInspection());
        myFixture.configureByFile("shed/pots.php");
        myFixture.testHighlighting(true, false, true);
    }
}
`

func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func fakeEA(t *testing.T) string {
	ea := t.TempDir()
	write(t, filepath.Join(ea, pluginXML), pluginBody)
	write(t, filepath.Join(ea, rulesMD), rulesBody)
	write(t, filepath.Join(ea, javaMain, "org/fake/garden/AlphaInspection.java"), alphaSrc)
	write(t, filepath.Join(ea, javaMain, "org/fake/BetaInspection.java"), "class BetaInspection { Object f = new TrimFix(); }\n")
	write(t, filepath.Join(ea, javaMain, "org/fake/garden/GammaInspection.java"), "class GammaInspection {}\n")
	write(t, filepath.Join(ea, javaMain, "org/fake/DeltaInspection.java"), "class DeltaInspection {}\n")
	write(t, filepath.Join(ea, descrDir, "AlphaInspection.html"), "<p>invented</p>\n")
	write(t, filepath.Join(ea, javaTest, "org/fake/garden/GardenTest.java"), gardenTest)
	write(t, filepath.Join(ea, javaTest, "org/fake/ShedTest.java"), shedTest)
	write(t, filepath.Join(ea, javaTest, "org/fake/notes.txt"), "public void testIgnored() {}\n")
	return ea
}

func TestRun(t *testing.T) {
	ea := fakeEA(t)
	root := t.TempDir()
	var stdout, stderr bytes.Buffer
	if code := run(root, []string{"-ea", ea}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr.String())
	}
	wantOut := strings.Join([]string{
		"rules: 4 (with fix: 2)",
		"cases: 5 (with .fixed: 2), rules covered: 3",
		"  no cases: Delta",
		"  unresolved: src/test/java/org/fake/garden/GardenTest.java#testBasics: new Rake()",
		"  unresolved: src/test/java/org/fake/garden/GardenTest.java#testBasics: quietAlpha",
		"  unresolved: src/test/java/org/fake/garden/GardenTest.java#testBasics: nothing()",
		"  unresolved: src/test/java/org/fake/garden/GardenTest.java#testBasics: stray",
		"",
	}, "\n")
	if stdout.String() != wantOut {
		t.Errorf("stdout:\n%s\nwant:\n%s", stdout.String(), wantOut)
	}

	var rules []meta.Rule
	readJSON(t, filepath.Join(root, metaOut), &rules)
	wantRules := []meta.Rule{
		{ID: "Alpha", LegacyID: "AlphaInspection", Group: "Garden", Severity: diagnostic.SeverityWarning, EnabledByDefault: true, HasFix: true, Options: []meta.Option{
			{Name: "LOUD", Type: "bool", Default: true},
			{Name: "LIMIT", Type: "int", Default: float64(5)},
			{Name: "ODD", Type: "int"},
			{Name: "LABEL", Type: "string", Default: "sprout"},
			{Name: "FLAVOUR", Type: "enum", Default: "PHPUNIT42"},
			{Name: "WORDS", Type: "list"},
		}},
		{ID: "Beta", LegacyID: "BetaInspection", Group: "Shed", Severity: diagnostic.SeverityError, HasFix: true},
		{ID: "Delta", LegacyID: "DeltaInspection", Group: "Shed", Severity: diagnostic.SeverityInfo, Experimental: true},
		{ID: "Gamma", LegacyID: "GammaInspection", Group: "Garden", Severity: diagnostic.SeverityInfo, EnabledByDefault: true},
	}
	if !reflect.DeepEqual(rules, wantRules) {
		t.Errorf("rules:\n%+v\nwant:\n%+v", rules, wantRules)
	}

	var idx Index
	readJSON(t, filepath.Join(root, indexOut), &idx)
	if idx.EAPath != ea {
		t.Errorf("eaPath = %q", idx.EAPath)
	}
	garden := "src/test/java/org/fake/garden/GardenTest.java"
	shed := "src/test/java/org/fake/ShedTest.java"
	wantIdx := map[string]RuleIndex{
		"Alpha": {LegacyID: "AlphaInspection", Class: "org.fake.garden.AlphaInspection", Source: "src/main/java/org/fake/garden/AlphaInspection.java", Description: "src/main/resources/inspectionDescriptions/AlphaInspection.html", Tests: []string{garden}},
		"Beta":  {LegacyID: "BetaInspection", Class: "org.fake.BetaInspection", Source: "src/main/java/org/fake/BetaInspection.java", Tests: []string{shed, garden}},
		"Gamma": {LegacyID: "GammaInspection", Class: "org.fake.garden.GammaInspection", Source: "src/main/java/org/fake/garden/GammaInspection.java", Tests: []string{garden}},
		"Delta": {LegacyID: "DeltaInspection", Class: "org.fake.DeltaInspection", Source: "src/main/java/org/fake/DeltaInspection.java"},
	}
	if !reflect.DeepEqual(idx.Rules, wantIdx) {
		t.Errorf("rules index:\n%+v\nwant:\n%+v", idx.Rules, wantIdx)
	}
	wantCases := []Case{
		{Test: shed + "#testPots", Rules: []string{"Beta"}, Fixture: "shed/pots.php"},
		{
			Test: garden + "#testBasics", Rules: []string{"Alpha", "Beta", "Alpha"}, Fixture: "garden/basic.php",
			Fixed: "garden/basic.fixed.php", PHP: "7.4", ComparisonStyle: "yoda",
			Options: map[string]string{"Alpha.LIMIT": "9", "Alpha.LOUD": "false"},
			Calls:   []string{`Alpha.addWord("ivy")`, `Alpha.addWord("moss")`},
		},
		{Test: garden + "#testLevels", Rules: []string{"Gamma"}, Fixture: "garden/first.php", PHP: "7.1"}, // snapshot taken at the next configureByFile
		{Test: garden + "#testLevels", Rules: []string{"Gamma"}, Fixture: "garden/second.php", PHP: "7.1", Companions: []string{"garden/first.php"}},
		{Test: garden + "#testLevels", Rules: []string{"Gamma"}, Fixture: "garden/third.php", Fixed: "garden/third.fixed.php", PHP: "5.6", Companions: []string{"garden/first.php", "garden/second.php"}},
	}
	// Walk order is lexical: org/fake/ShedTest.java before org/fake/garden/.
	if len(idx.Cases) != len(wantCases) {
		t.Fatalf("cases: %+v", idx.Cases)
	}
	for i, w := range wantCases {
		if !reflect.DeepEqual(idx.Cases[i], w) {
			t.Errorf("case %d:\n%+v\nwant:\n%+v", i, idx.Cases[i], w)
		}
	}
}

func readJSON(t *testing.T, path string, v any) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, v); err != nil {
		t.Fatal(err)
	}
}

func TestRunFlags(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run(t.TempDir(), []string{"-h"}, &stdout, &stderr); code != 0 {
		t.Errorf("-h: exit %d", code)
	}
	if code := run(t.TempDir(), []string{"-bogus"}, &stdout, &stderr); code != 2 {
		t.Errorf("bad flag: exit %d", code)
	}
}

func TestRunErrors(t *testing.T) {
	cases := map[string]struct {
		edit func(t *testing.T, ea, root string)
		want string
	}{
		"missing plugin.xml": {func(t *testing.T, ea, _ string) { os.Remove(filepath.Join(ea, pluginXML)) }, "no such file"},
		"malformed plugin.xml": {func(t *testing.T, ea, _ string) {
			write(t, filepath.Join(ea, pluginXML), "<idea-plugin><localInspection shortName=")
		}, "plugin.xml: "},
		"incomplete declaration": {func(t *testing.T, ea, _ string) {
			write(t, filepath.Join(ea, pluginXML), `<idea-plugin><localInspection shortName="OnlyInspection"/></idea-plugin>`)
		}, "incomplete localInspection"},
		"missing RULES.md": {func(t *testing.T, ea, _ string) { os.Remove(filepath.Join(ea, rulesMD)) }, "RULES.md"},
		"overlong RULES.md line": {func(t *testing.T, ea, _ string) {
			write(t, filepath.Join(ea, rulesMD), strings.Repeat("x", 70*1024)+"\n")
		}, "too long"},
		"duplicate id": {func(t *testing.T, ea, _ string) {
			write(t, filepath.Join(ea, pluginXML), `<a><localInspection shortName="AlphaInspection" level="ERROR" implementationClass="org.fake.garden.AlphaInspection"/><localInspection shortName="AlphaInspection" level="ERROR" implementationClass="org.fake.garden.AlphaInspection"/></a>`)
		}, "duplicate rule id Alpha"},
		"missing source": {func(t *testing.T, ea, _ string) {
			os.Remove(filepath.Join(ea, javaMain, "org/fake/BetaInspection.java"))
		}, "BetaInspection: "},
		"unknown level": {func(t *testing.T, ea, _ string) {
			write(t, filepath.Join(ea, pluginXML), `<a><localInspection shortName="AlphaInspection" level="LOUD" implementationClass="org.fake.garden.AlphaInspection"/></a>`)
		}, `unknown level "LOUD"`},
		"missing test tree": {func(t *testing.T, ea, _ string) { os.RemoveAll(filepath.Join(ea, javaTest)) }, "src/test/java"},
		"unreadable test": {func(t *testing.T, ea, _ string) {
			if err := os.Symlink(filepath.Join(ea, "nowhere"), filepath.Join(ea, javaTest, "BrokenTest.java")); err != nil {
				t.Fatal(err)
			}
		}, "BrokenTest.java"},
		"unwritable rules.json": {func(t *testing.T, _, root string) { write(t, filepath.Join(root, "internal"), "a file") }, "internal"},
		"unwritable index":      {func(t *testing.T, _, root string) { write(t, filepath.Join(root, ".cache"), "a file") }, ".cache"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			ea, root := fakeEA(t), t.TempDir()
			c.edit(t, ea, root)
			var stdout, stderr bytes.Buffer
			if code := run(root, []string{"-ea", ea}, &stdout, &stderr); code != 1 || !strings.HasPrefix(stderr.String(), "extract: ") || !strings.Contains(stderr.String(), c.want) {
				t.Fatalf("exit %d, stderr %q, want %q", code, stderr.String(), c.want)
			}
		})
	}
}

func TestAbsError(t *testing.T) {
	old := abs
	abs = func(string) (string, error) { return "", errors.New("no working directory") }
	t.Cleanup(func() { abs = old })
	var stdout, stderr bytes.Buffer
	if code := run(t.TempDir(), []string{"-ea", "relative"}, &stdout, &stderr); code != 1 || stderr.String() != "extract: no working directory\n" {
		t.Fatalf("exit %d, %q", code, stderr.String())
	}
}

func TestSeverity(t *testing.T) {
	for level, want := range map[string]diagnostic.Severity{
		"ERROR": diagnostic.SeverityError, "WARNING": diagnostic.SeverityWarning,
		"WEAK WARNING": diagnostic.SeverityInfo, "INFO": diagnostic.SeverityInfo, "INFORMATION": diagnostic.SeverityInfo, "TYPO": diagnostic.SeverityInfo,
	} {
		if got, err := severity(level); err != nil || got != want {
			t.Errorf("severity(%q) = %q, %v", level, got, err)
		}
	}
}

func TestSplitArgs(t *testing.T) {
	got := splitArgs(" a, new B(c(1, 2), d) ,e ")
	want := []string{"a", "new B(c(1, 2), d)", "e"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("splitArgs = %q", got)
	}
	if got := splitArgs("  "); got != nil {
		t.Errorf("splitArgs(blank) = %q", got)
	}
}
