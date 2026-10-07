package compatibility

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

// deprecatedIniOptions reports ini_* calls on directives deprecated or
// removed at the targeted PHP version.
type deprecatedIniOptions struct{}

func init() { register(deprecatedIniOptions{}) }

type iniDirective struct {
	deprecated, removed phpver.Version // 0 = none
	alt                 string
}

var deprecatedIni = map[string]iniDirective{
	"define_syslog_variables":         {phpver.PHP53, phpver.PHP54, ""},
	"magic_quotes_gpc":                {phpver.PHP53, phpver.PHP54, ""},
	"magic_quotes_runtime":            {phpver.PHP53, phpver.PHP54, ""},
	"magic_quotes_sybase":             {phpver.PHP53, phpver.PHP54, ""},
	"highlight.bg":                    {0, phpver.PHP54, ""},
	"xsl.security_prefs":              {phpver.PHP54, phpver.PHP70, "XsltProcessor->setSecurityPrefs()"},
	"safe_mode":                       {phpver.PHP53, phpver.PHP54, ""},
	"safe_mode_gid":                   {phpver.PHP53, phpver.PHP54, ""},
	"safe_mode_include_dir":           {phpver.PHP53, phpver.PHP54, ""},
	"safe_mode_exec_dir":              {phpver.PHP53, phpver.PHP54, ""},
	"safe_mode_allowed_env_vars":      {phpver.PHP53, phpver.PHP54, ""},
	"safe_mode_protected_env_vars":    {phpver.PHP53, phpver.PHP54, ""},
	"sql.safe_mode":                   {0, phpver.PHP72, ""},
	"asp_tags":                        {0, phpver.PHP70, ""},
	"always_populate_raw_post_data":   {phpver.PHP56, phpver.PHP70, ""},
	"y2k_compliance":                  {0, phpver.PHP54, ""},
	"zend.ze1_compatibility_mode":     {0, phpver.PHP53, ""},
	"allow_call_time_pass_reference":  {phpver.PHP53, phpver.PHP54, ""},
	"register_globals":                {phpver.PHP53, phpver.PHP54, ""},
	"register_long_arrays":            {phpver.PHP53, phpver.PHP54, ""},
	"session.hash_function":           {0, phpver.PHP71, ""},
	"session.hash_bits_per_character": {0, phpver.PHP71, ""},
	"session.entropy_file":            {0, phpver.PHP71, ""},
	"session.entropy_length":          {0, phpver.PHP71, ""},
	"session.bug_compat_42":           {0, phpver.PHP54, ""},
	"session.bug_compat_warn":         {0, phpver.PHP54, ""},
	"iconv.input_encoding":            {phpver.PHP56, 0, "default_charset"},
	"iconv.output_encoding":           {phpver.PHP56, 0, "default_charset"},
	"iconv.internal_encoding":         {phpver.PHP56, 0, "default_charset"},
	"mbstring.script_encoding":        {0, phpver.PHP54, "zend.script_encoding"},
	"mbstring.func_overload":          {phpver.PHP72, phpver.PHP80, ""},
	"mbstring.internal_encoding":      {phpver.PHP56, 0, "default_charset"},
	"mbstring.http_input":             {phpver.PHP56, 0, "default_charset"},
	"mbstring.http_output":            {phpver.PHP56, 0, "default_charset"},
	"track_errors":                    {phpver.PHP72, phpver.PHP80, ""},
	"pdo_odbc.db2_instance_name":      {phpver.PHP73, phpver.PHP80, ""},
	"opcache.load_comments":           {0, phpver.PHP70, ""},
	"opcache.fast_shutdown":           {0, phpver.PHP72, ""},
	"opcache.inherited_hack":          {0, phpver.PHP73, ""},
	// Later PHP versions (custos extension, spec Divergences).
	"allow_url_include":              {phpver.PHP74, 0, ""},
	"assert.quiet_eval":              {0, phpver.PHP80, ""},
	"log_errors_max_len":             {0, phpver.PHP81, ""},
	"mysqlnd.fetch_data_copy":        {0, phpver.PHP81, ""},
	"auto_detect_line_endings":       {phpver.PHP81, 0, ""},
	"date.default_latitude":          {phpver.PHP81, 0, ""},
	"date.default_longitude":         {phpver.PHP81, 0, ""},
	"date.sunrise_zenith":            {phpver.PHP81, 0, ""},
	"date.sunset_zenith":             {phpver.PHP81, 0, ""},
	"filter.default":                 {phpver.PHP81, 0, ""},
	"filter.default_flags":           {phpver.PHP81, 0, ""},
	"assert.active":                  {phpver.PHP83, 0, "zend.assertions"},
	"assert.bail":                    {phpver.PHP83, 0, ""},
	"assert.callback":                {phpver.PHP83, 0, ""},
	"assert.exception":               {phpver.PHP83, 0, ""},
	"assert.warning":                 {phpver.PHP83, 0, ""},
	"opcache.consistency_checks":     {0, phpver.PHP83, ""},
	"session.sid_length":             {phpver.PHP84, 0, ""},
	"session.sid_bits_per_character": {phpver.PHP84, 0, ""},
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
	switch ctx.GlobalFunctionName(call) { // D1
	case "ini_set", "ini_get", "ini_alter", "ini_restore":
	default:
		return
	}
	arg, ok := call.Args.Args[0].(*syntax.Arg)
	if !ok || arg.Value == nil {
		return
	}
	raw, ok := util.QuotedStringContent(arg.Value) // D2
	if !ok {
		return
	}
	directive := strings.ToLower(raw)
	d, ok := deprecatedIni[directive] // D3
	if !ok {
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

func fullVersion(v phpver.Version) string { return v.String() + ".0" }
