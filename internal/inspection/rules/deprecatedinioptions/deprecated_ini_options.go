package deprecatedinioptions

import (
	"strconv"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// deprecatedIniOptions reports ini_* calls on directives deprecated or
// removed at the targeted PHP phpversion.
type (
	deprecatedIniOptions struct{}
	iniDirective         struct {
		deprecated, removed phpversion.Version // 0 = none
		alt                 string
	}
)

var deprecatedIni = map[string]iniDirective{
	"define_syslog_variables":         {phpversion.PHP53, phpversion.PHP54, ""},
	"magic_quotes_gpc":                {phpversion.PHP53, phpversion.PHP54, ""},
	"magic_quotes_runtime":            {phpversion.PHP53, phpversion.PHP54, ""},
	"magic_quotes_sybase":             {phpversion.PHP53, phpversion.PHP54, ""},
	"highlight.bg":                    {0, phpversion.PHP54, ""},
	"xsl.security_prefs":              {phpversion.PHP54, phpversion.PHP70, "XsltProcessor->setSecurityPrefs()"},
	"safe_mode":                       {phpversion.PHP53, phpversion.PHP54, ""},
	"safe_mode_gid":                   {phpversion.PHP53, phpversion.PHP54, ""},
	"safe_mode_include_dir":           {phpversion.PHP53, phpversion.PHP54, ""},
	"safe_mode_exec_dir":              {phpversion.PHP53, phpversion.PHP54, ""},
	"safe_mode_allowed_env_vars":      {phpversion.PHP53, phpversion.PHP54, ""},
	"safe_mode_protected_env_vars":    {phpversion.PHP53, phpversion.PHP54, ""},
	"sql.safe_mode":                   {0, phpversion.PHP72, ""},
	"asp_tags":                        {0, phpversion.PHP70, ""},
	"always_populate_raw_post_data":   {phpversion.PHP56, phpversion.PHP70, ""},
	"y2k_compliance":                  {0, phpversion.PHP54, ""},
	"zend.ze1_compatibility_mode":     {0, phpversion.PHP53, ""},
	"allow_call_time_pass_reference":  {phpversion.PHP53, phpversion.PHP54, ""},
	"register_globals":                {phpversion.PHP53, phpversion.PHP54, ""},
	"register_long_arrays":            {phpversion.PHP53, phpversion.PHP54, ""},
	"session.hash_function":           {0, phpversion.PHP71, ""},
	"session.hash_bits_per_character": {0, phpversion.PHP71, ""},
	"session.entropy_file":            {0, phpversion.PHP71, ""},
	"session.entropy_length":          {0, phpversion.PHP71, ""},
	"session.bug_compat_42":           {0, phpversion.PHP54, ""},
	"session.bug_compat_warn":         {0, phpversion.PHP54, ""},
	"iconv.input_encoding":            {phpversion.PHP56, 0, "default_charset"},
	"iconv.output_encoding":           {phpversion.PHP56, 0, "default_charset"},
	"iconv.internal_encoding":         {phpversion.PHP56, 0, "default_charset"},
	"mbstring.script_encoding":        {0, phpversion.PHP54, "zend.script_encoding"},
	"mbstring.func_overload":          {phpversion.PHP72, phpversion.PHP80, ""},
	"mbstring.internal_encoding":      {phpversion.PHP56, 0, "default_charset"},
	"mbstring.http_input":             {phpversion.PHP56, 0, "default_charset"},
	"mbstring.http_output":            {phpversion.PHP56, 0, "default_charset"},
	"track_errors":                    {phpversion.PHP72, phpversion.PHP80, ""},
	"pdo_odbc.db2_instance_name":      {phpversion.PHP73, phpversion.PHP80, ""},
	"opcache.load_comments":           {0, phpversion.PHP70, ""},
	"opcache.fast_shutdown":           {0, phpversion.PHP72, ""},
	"opcache.inherited_hack":          {0, phpversion.PHP73, ""},
	// Later PHP versions (custos extension, spec Divergences).
	"allow_url_include":              {phpversion.PHP74, 0, ""},
	"assert.quiet_eval":              {0, phpversion.PHP80, ""},
	"log_errors_max_len":             {0, phpversion.PHP81, ""},
	"mysqlnd.fetch_data_copy":        {0, phpversion.PHP81, ""},
	"auto_detect_line_endings":       {phpversion.PHP81, 0, ""},
	"date.default_latitude":          {phpversion.PHP81, 0, ""},
	"date.default_longitude":         {phpversion.PHP81, 0, ""},
	"date.sunrise_zenith":            {phpversion.PHP81, 0, ""},
	"date.sunset_zenith":             {phpversion.PHP81, 0, ""},
	"filter.default":                 {phpversion.PHP81, 0, ""},
	"filter.default_flags":           {phpversion.PHP81, 0, ""},
	"assert.active":                  {phpversion.PHP83, 0, "zend.assertions"},
	"assert.bail":                    {phpversion.PHP83, 0, ""},
	"assert.callback":                {phpversion.PHP83, 0, ""},
	"assert.exception":               {phpversion.PHP83, 0, ""},
	"assert.warning":                 {phpversion.PHP83, 0, ""},
	"opcache.consistency_checks":     {0, phpversion.PHP83, ""},
	"session.sid_length":             {phpversion.PHP84, 0, ""},
	"session.sid_bits_per_character": {phpversion.PHP84, 0, ""},
}

// defaultValued holds the directives PHP only deprecates when set to a value
// other than their default: whether a written value (as ini_set() receives
// it) is that default.
var defaultValued = map[string]func(string) bool{
	"assert.active":                  iniBoolIs(true),
	"assert.bail":                    iniBoolIs(false),
	"assert.callback":                func(v string) bool { return v == "" },
	"assert.exception":               iniBoolIs(true),
	"assert.warning":                 iniBoolIs(true),
	"session.sid_length":             iniIntIs(32),
	"session.sid_bits_per_character": iniIntIs(4),
}

func iniBoolIs(want bool) func(string) bool {
	return func(v string) bool {
		switch strings.ToLower(v) {
		case "on", "yes", "true":
			return want
		case "", "off", "no", "false", "none":
			return !want
		}
		n, err := strconv.Atoi(v)
		return err == nil && (n != 0) == want
	}
}

func iniIntIs(want int) func(string) bool {
	return func(v string) bool {
		n, err := strconv.Atoi(v)
		return err == nil && n == want
	}
}

// iniValue returns the string ini_set() receives for a literal value.
func iniValue(e syntax.Expr) (string, bool) {
	if s, ok := astquery.QuotedStringContent(e); ok {
		return s, !strings.Contains(s, `\`)
	}
	if b, ok := astquery.BoolConst(e); ok {
		if b {
			return "1", true
		}
		return "", true
	}
	switch x := e.(type) {
	case *syntax.ConstFetch:
		return "", strings.EqualFold(strings.TrimPrefix(x.Name.Value, `\`), "null")
	case *syntax.Literal:
		if n, ok := astquery.ParseIntLiteral(x.Raw); ok && x.LitKind == syntax.LitInt {
			return strconv.FormatInt(n, 10), true
		}
	}
	return "", false
}

// keepsDefault reports a read, a restore or an assignment of the default
// value of a directive deprecated only for other values (E5).
func keepsDefault(fn, directive string, call *syntax.FuncCall) bool {
	isDefault, ok := defaultValued[directive]
	if !ok {
		return false
	}
	if fn == "ini_get" || fn == "ini_restore" {
		return true
	}
	if len(call.Args.Args) < 2 {
		return false
	}
	arg, ok := call.Args.Args[1].(*syntax.Arg)
	if !ok || arg.Unpack || arg.Name != nil && arg.Name.Value != "value" {
		return false
	}
	v, ok := iniValue(arg.Value)
	return ok && isDefault(v)
}

func (deprecatedIniOptions) ID() string { return "DeprecatedIniOptions" }
func (deprecatedIniOptions) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall}
}

func (deprecatedIniOptions) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if call.Args == nil || len(call.Args.Args) == 0 {
		return
	}
	fn := ctx.GlobalFunctionName(call)
	switch fn { // D1
	case "ini_set", "ini_get", "ini_alter", "ini_restore":
	default:
		return
	}
	arg, ok := call.Args.Args[0].(*syntax.Arg)
	if !ok || arg.Value == nil {
		return
	}
	raw, ok := astquery.QuotedStringContent(arg.Value) // D2
	if !ok {
		return
	}
	directive := strings.ToLower(raw)
	d, ok := deprecatedIni[directive] // D3
	if !ok || keepsDefault(fn, directive, call) {
		return
	}
	var msg string
	switch { // D4
	case d.removed != 0 && ctx.PHP.AtLeast(d.removed):
		msg = "Ini directive '" + directive + "' no longer exists since PHP " + fullVersion(d.removed)
	case d.deprecated != 0 && ctx.PHP.AtLeast(d.deprecated):
		msg = "Ini directive '" + directive + "' is deprecated since PHP " + fullVersion(d.deprecated)
	default:
		return
	}
	if d.alt != "" {
		msg += "; use " + d.alt
	}
	ctx.ReportNode(arg.Value, msg+".")
}
func fullVersion(v phpversion.Version) string { return v.String() + ".0" }
