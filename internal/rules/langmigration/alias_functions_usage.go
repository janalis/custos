package langmigration

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// aliasFunctionsUsage reports calls to built-in alias functions (`sizeof`,
// `join`, …) and offers the canonical name.
type aliasFunctionsUsage struct{}

func init() { register(aliasFunctionsUsage{}) }

func (aliasFunctionsUsage) ID() string { return "AliasFunctionsUsage" }

func (aliasFunctionsUsage) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

// aliasCanonical is table A: replaceable aliases.
var aliasCanonical = map[string]string{
	"close":                   "closedir",
	"is_double":               "is_float",
	"is_integer":              "is_int",
	"is_long":                 "is_int",
	"is_real":                 "is_float",
	"sizeof":                  "count",
	"doubleval":               "floatval",
	"fputs":                   "fwrite",
	"join":                    "implode",
	"key_exists":              "array_key_exists",
	"chop":                    "rtrim",
	"ini_alter":               "ini_set",
	"is_writeable":            "is_writable",
	"pos":                     "current",
	"show_source":             "highlight_file",
	"strchr":                  "strstr",
	"set_file_buffer":         "stream_set_write_buffer",
	"session_commit":          "session_write_close",
	"socket_getopt":           "socket_get_option",
	"socket_setopt":           "socket_set_option",
	"openssl_get_privatekey":  "openssl_pkey_get_private",
	"posix_errno":             "posix_get_last_error",
	"ldap_close":              "ldap_unbind",
	"pcntl_errno":             "pcntl_get_last_error",
	"ftp_quit":                "ftp_close",
	"socket_set_blocking":     "stream_set_blocking",
	"stream_register_wrapper": "stream_wrapper_register",
	"socket_set_timeout":      "stream_set_timeout",
	"socket_get_status":       "stream_get_meta_data",
	"diskfreespace":           "disk_free_space",
	"odbc_do":                 "odbc_exec",
	"odbc_field_precision":    "odbc_field_len",
	"recode":                  "recode_string",
	"mysqli_escape_string":    "mysqli_real_escape_string",
	"mysqli_execute":          "mysqli_stmt_execute",
}

// aliasLegacy is table B: legacy aliases, reported without a fix.
var aliasLegacy = map[string]string{
	"mysqli_bind_param":      "deprecated 5.3, removed 5.4",
	"mysqli_bind_result":     "deprecated 5.3, removed 5.4",
	"mysqli_client_encoding": "deprecated 5.3, removed 5.4",
	"mysqli_fetch":           "deprecated 5.3, removed 5.4",
	"mysqli_param_count":     "deprecated 5.3, removed 5.4",
	"mysqli_get_metadata":    "deprecated 5.3, removed 5.4",
	"mysqli_send_long_data":  "deprecated 5.3, removed 5.4",
	"ocifreecursor":          "deprecated 5.4, discouraged",
	"magic_quotes_runtime":   "deprecated 5.3, removed 7.0",
}

func (aliasFunctionsUsage) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	name, part, ok := util.FuncNamePart(call)
	if !ok {
		return
	}
	key := strings.ToLower(part) // D1: PHP function names are case-insensitive
	canonical, replaceable := aliasCanonical[key]
	status, legacy := aliasLegacy[key]
	if !replaceable && !legacy {
		return
	}
	if !callsGlobalFunction(ctx, call, false) { // D2, E1
		return
	}
	span := util.NamePartSpan(name)
	if legacy {
		ctx.Report(span, "'"+part+"(...)' is a legacy alias ("+status+"); stop relying on it.")
		return
	}
	if !strings.Contains(name.Value, `\`) { // a namespaced function of the canonical name would capture a bare call
		canonical = util.QualifiedBuiltin(ctx, canonical, call.Span().Start)
	}
	ctx.Report(span, "Use '"+canonical+"(...)' instead of the alias '"+part+"(...)'.", analysis.Fix{
		Title: "Use '" + canonical + "(...)'",
		Edits: func() []analysis.TextEdit {
			return []analysis.TextEdit{{Span: span, NewText: canonical}}
		},
	})
}
