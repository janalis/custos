package syntax

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	phpversion "custos/internal/php/version"
)

// phpTokenKinds maps PHP's token_name() output to our kinds.
var phpTokenKinds = map[string]TokenKind{
	"T_WHITESPACE": TWhitespace, "T_COMMENT": TComment, "T_DOC_COMMENT": TDocComment,
	"T_INLINE_HTML": TInlineHTML, "T_OPEN_TAG": TOpenTag, "T_OPEN_TAG_WITH_ECHO": TOpenTagWithEcho,
	"T_CLOSE_TAG": TCloseTag, "T_VARIABLE": TVariable, "T_STRING": TString,
	"T_NAME_QUALIFIED": TNameQualified, "T_NAME_FULLY_QUALIFIED": TNameFullyQualified,
	"T_NAME_RELATIVE": TNameRelative, "T_LNUMBER": TLNumber, "T_DNUMBER": TDNumber,
	"T_CONSTANT_ENCAPSED_STRING": TConstantEncapsedString, "T_ENCAPSED_AND_WHITESPACE": TEncapsedAndWhitespace,
	"T_STRING_VARNAME": TStringVarname, "T_NUM_STRING": TNumString, "T_START_HEREDOC": TStartHeredoc,
	"T_END_HEREDOC": TEndHeredoc, "T_DOLLAR_OPEN_CURLY_BRACES": TDollarOpenCurlyBraces,
	"T_CURLY_OPEN": TCurlyOpen, "T_BAD_CHARACTER": TBadCharacter,
	"T_INT_CAST": TIntCast, "T_DOUBLE_CAST": TDoubleCast, "T_STRING_CAST": TStringCast,
	"T_ARRAY_CAST": TArrayCast, "T_OBJECT_CAST": TObjectCast, "T_BOOL_CAST": TBoolCast,
	"T_UNSET_CAST": TUnsetCast, "T_VOID_CAST": TVoidCast,
	"T_OBJECT_OPERATOR": TObjectOperator, "T_NULLSAFE_OBJECT_OPERATOR": TNullsafeObjectOperator,
	"T_DOUBLE_ARROW": TDoubleArrow, "T_DOUBLE_COLON": TPaamayimNekudotayim, "T_PAAMAYIM_NEKUDOTAYIM": TPaamayimNekudotayim,
	"T_NS_SEPARATOR": TNsSeparator, "T_ELLIPSIS": TEllipsis, "T_COALESCE": TCoalesce,
	"T_COALESCE_EQUAL": TCoalesceEqual, "T_POW": TPow, "T_POW_EQUAL": TPowEqual, "T_INC": TInc, "T_DEC": TDec,
	"T_IS_EQUAL": TIsEqual, "T_IS_NOT_EQUAL": TIsNotEqual, "T_IS_IDENTICAL": TIsIdentical,
	"T_IS_NOT_IDENTICAL": TIsNotIdentical, "T_IS_SMALLER_OR_EQUAL": TIsSmallerOrEqual,
	"T_IS_GREATER_OR_EQUAL": TIsGreaterOrEqual, "T_SPACESHIP": TSpaceship, "T_PLUS_EQUAL": TPlusEqual,
	"T_MINUS_EQUAL": TMinusEqual, "T_MUL_EQUAL": TMulEqual, "T_DIV_EQUAL": TDivEqual,
	"T_CONCAT_EQUAL": TConcatEqual, "T_MOD_EQUAL": TModEqual, "T_AND_EQUAL": TAndEqual, "T_OR_EQUAL": TOrEqual,
	"T_XOR_EQUAL": TXorEqual, "T_SL_EQUAL": TSlEqual, "T_SR_EQUAL": TSrEqual, "T_SL": TSl, "T_SR": TSr,
	"T_BOOLEAN_OR": TBooleanOr, "T_BOOLEAN_AND": TBooleanAnd, "T_PIPE": TPipe, "T_ATTRIBUTE": TAttribute,
	"T_AMPERSAND_FOLLOWED_BY_VAR_OR_VARARG": TAmpersand, "T_AMPERSAND_NOT_FOLLOWED_BY_VAR_OR_VARARG": TAmpersand,
	"T_LOGICAL_AND": TAnd, "T_LOGICAL_OR": TOr, "T_LOGICAL_XOR": TXor,
	"T_ABSTRACT": TAbstract, "T_ARRAY": TArray, "T_AS": TAs, "T_BREAK": TBreak, "T_CALLABLE": TCallable,
	"T_CASE": TCase, "T_CATCH": TCatch, "T_CLASS": TClass, "T_CLONE": TClone, "T_CONST": TConst,
	"T_CONTINUE": TContinue, "T_DECLARE": TDeclare, "T_DEFAULT": TDefault, "T_DO": TDo, "T_ECHO": TEcho,
	"T_ELSE": TElse, "T_ELSEIF": TElseif, "T_EMPTY": TEmpty, "T_ENDDECLARE": TEnddeclare, "T_ENDFOR": TEndfor,
	"T_ENDFOREACH": TEndforeach, "T_ENDIF": TEndif, "T_ENDSWITCH": TEndswitch, "T_ENDWHILE": TEndwhile,
	"T_ENUM": TEnum, "T_EVAL": TEval, "T_EXIT": TExit, "T_EXTENDS": TExtends, "T_FINAL": TFinal,
	"T_FINALLY": TFinally, "T_FN": TFn, "T_FOR": TFor, "T_FOREACH": TForeach, "T_FUNCTION": TFunction,
	"T_GLOBAL": TGlobal, "T_GOTO": TGoto, "T_IF": TIf, "T_IMPLEMENTS": TImplements, "T_INCLUDE": TInclude,
	"T_INCLUDE_ONCE": TIncludeOnce, "T_INSTANCEOF": TInstanceof, "T_INSTEADOF": TInsteadof,
	"T_INTERFACE": TInterface, "T_ISSET": TIsset, "T_LIST": TList, "T_MATCH": TMatch,
	"T_NAMESPACE": TNamespace, "T_NEW": TNew, "T_PRINT": TPrint, "T_PRIVATE": TPrivate,
	"T_PROTECTED": TProtected, "T_PUBLIC": TPublic, "T_PRIVATE_SET": TPrivateSet,
	"T_PROTECTED_SET": TProtectedSet, "T_PUBLIC_SET": TPublicSet, "T_READONLY": TReadonly,
	"T_REQUIRE": TRequire, "T_REQUIRE_ONCE": TRequireOnce, "T_RETURN": TReturn, "T_STATIC": TStatic,
	"T_SWITCH": TSwitch, "T_THROW": TThrow, "T_TRAIT": TTrait, "T_TRY": TTry, "T_UNSET": TUnset,
	"T_USE": TUse, "T_VAR": TVar, "T_WHILE": TWhile, "T_YIELD": TYield, "T_YIELD_FROM": TYieldFrom,
	"T_HALT_COMPILER": THaltCompiler, "T_LINE": TLine, "T_FILE": TFile, "T_DIR": TDir,
	"T_CLASS_C": TClassC, "T_TRAIT_C": TTraitC, "T_METHOD_C": TMethodC, "T_FUNC_C": TFuncC,
	"T_NS_C": TNsC, "T_PROPERTY_C": TPropertyC,
}

var phpCharKinds = map[string]TokenKind{
	";": TSemicolon, ",": TComma, ".": TDot, "{": TLBrace, "}": TRBrace, "(": TLParen, ")": TRParen,
	"[": TLBracket, "]": TRBracket, "+": TPlus, "-": TMinus, "*": TMul, "/": TDiv, "%": TMod, "=": TEqual,
	"<": TLess, ">": TGreater, "!": TExclaim, "?": TQuestion, ":": TColon, "&": TAmpersand, "|": TBar,
	"^": TCaret, "~": TTilde, "@": TAt, "$": TDollar, `"`: TDoubleQuote, "`": TBacktick, `b"`: TDoubleQuote, `B"`: TDoubleQuote,
}

type oracleTok struct {
	kind       TokenKind
	name       string
	start, end int
}

// corpusFiles lists PHP files under CUSTOS_CORPUS.
func corpusFiles(t *testing.T) []string {
	t.Helper()
	root := os.Getenv("CUSTOS_CORPUS")
	if root == "" {
		t.Skip("set CUSTOS_CORPUS")
	}
	if _, err := os.Stat(root); err != nil {
		t.Skipf("corpus %s not found", root)
	}
	limit := 4000
	if s := os.Getenv("CUSTOS_CORPUS_MAX"); s != "" {
		limit, _ = strconv.Atoi(s)
	}
	var files []string
	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || len(files) >= limit {
			return filepath.SkipDir
		}
		if !d.IsDir() && strings.HasSuffix(p, ".php") {
			files = append(files, p)
		}
		return nil
	})
	return files
}

func runOracle(t *testing.T, files []string) map[string][]oracleTok {
	t.Helper()
	if _, err := exec.LookPath("php"); err != nil {
		t.Skip("php not installed")
	}
	cmd := exec.Command("php", "-d", "short_open_tag=0", "-d", "memory_limit=-1", "testdata/oracle/tokens.php")
	cmd.Stdin = strings.NewReader(strings.Join(files, "\n") + "\n")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("php oracle: %v", err)
	}
	res := map[string][]oracleTok{}
	var cur string
	sc := bufio.NewScanner(&out)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "== ") {
			cur = line[3:]
			res[cur] = nil
			continue
		}
		i := strings.LastIndexByte(line, ' ')
		j := strings.LastIndexByte(line[:i], ' ')
		name := line[:j]
		start, _ := strconv.Atoi(line[j+1 : i])
		end, _ := strconv.Atoi(line[i+1:])
		var k TokenKind
		var ok bool
		if strings.HasPrefix(name, "CHAR:") {
			k, ok = phpCharKinds[name[5:]]
		} else {
			k, ok = phpTokenKinds[name]
		}
		if !ok {
			k = TBadCharacter
		}
		res[cur] = append(res[cur], oracleTok{k, name, start, end})
	}
	return res
}

// TestLexerOracle compares our token stream with PHP's token_get_all on a
// corpus (PHP 8.5 semantics). Run with -short to skip.
func TestLexerOracle(t *testing.T) {
	if testing.Short() {
		t.Skip("short mode")
	}
	files := corpusFiles(t)
	oracle := runOracle(t, files)
	bad := 0
	for _, f := range files {
		want, ok := oracle[f]
		if !ok {
			continue
		}
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		got, _ := Lex(src, LexOptions{Version: phpversion.PHP85})
		if msg := compareTokens(src, got, want); msg != "" {
			bad++
			if bad <= 15 {
				t.Errorf("%s: %s", f, msg)
			}
		}
	}
	t.Logf("%d files, %d mismatching", len(files), bad)
	if bad > 15 {
		t.Errorf("... %d more mismatching files", bad-15)
	}
}

func compareTokens(src []byte, got []Token, want []oracleTok) string {
	// PHP returns __halt_compiler data as T_INLINE_HTML.
	n := len(got)
	if len(want) < n {
		n = len(want)
	}
	for i := 0; i < n; i++ {
		g, w := got[i], want[i]
		wk := w.kind
		if g.Kind == THaltCompilerData && w.name == "T_INLINE_HTML" {
			wk = THaltCompilerData
		}
		if g.Kind != wk || int(g.Start) != w.start || int(g.End) != w.end {
			ctx := src[max(0, w.start-30):min(len(src), w.end+30)]
			return fmt.Sprintf("token %d: got %s[%d:%d] want %s[%d:%d] near %q", i, g.Kind, g.Start, g.End, w.name, w.start, w.end, ctx)
		}
	}
	if len(got) != len(want) {
		return fmt.Sprintf("token count: got %d want %d", len(got), len(want))
	}
	return ""
}
