package aliasfunctionsusage

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// aliasFunctionsUsage reports calls to built-in alias functions (`sizeof`,
// `join`, …) and offers the canonical name.
type aliasFunctionsUsage struct{}

func (aliasFunctionsUsage) ID() string               { return "AliasFunctionsUsage" }
func (aliasFunctionsUsage) Semantic()                {}
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

// aliasLegacy is table B: legacy aliases, reported without a fixing.
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
	name, part, ok := astquery.FuncNamePart(call)
	if !ok {
		return
	}
	key := strings.ToLower(part) // D1: PHP function names are case-insensitive
	canonical, replaceable := aliasCanonical[key]
	status, legacy := aliasLegacy[key]
	if !replaceable && !legacy {
		return
	}
	global := ctx.GlobalFunctionName(call)
	if global == "" { // D2, E1
		return
	}
	// E5: PHP cannot redeclare a built-in, so a project declaration the index
	// resolves at this level is the function that runs, not the alias.
	fn := ctx.Index().Function(global, ctx.PHP)
	if fn != nil && !fn.Builtin {
		return
	}
	span := astquery.NamePartSpan(name)
	if legacy {
		ctx.Report(span, "'"+part+"(...)' is a legacy alias ("+status+"); stop relying on it.")
		return
	}
	message := "Use '" + canonical + "(...)' instead of the alias '" + part + "(...)'."
	// F3: where PHP lacks the alias, the call can only reach a function the
	// index does not see; renaming it would change what runs.
	if fn == nil || !fn.Avail.In(ctx.PHP) || removedBy(key, ctx.PHP) {
		ctx.Report(span, message)
		return
	}
	if !strings.Contains(name.Value, `\`) { // a namespaced function of the canonical name would capture a bare call
		canonical = semanticquery.QualifiedBuiltin(ctx, canonical, call.Span().Start)
		message = "Use '" + canonical + "(...)' instead of the alias '" + part + "(...)'."
	}
	ctx.Report(span, message, diagnostic.Fix{
		Title: "Use '" + canonical + "(...)'",
		Edits: func() []diagnostic.TextEdit {
			return []diagnostic.TextEdit{{Span: span, NewText: canonical}}
		},
	})
}

// aliasRemoved lists table A aliases PHP dropped while the stubs still
// declare them for every version, with the first version lacking them.
var aliasRemoved = map[string]phpversion.Version{
	"is_real": phpversion.PHP80,
}

// removedBy reports whether PHP no longer provides the alias at ver.
func removedBy(alias string, ver phpversion.Version) bool {
	since, ok := aliasRemoved[alias]
	return ok && ver >= since
}
