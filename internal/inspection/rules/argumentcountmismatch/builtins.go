package argumentcountmismatch

import (
	"strings"

	"github.com/janalis/custos/internal/inspection/semanticquery"
	phpversion "github.com/janalis/custos/internal/php/version"
)

// driftedStubs lists the builtins of bundled directories whose stub
// declares other parameter counts than PHP does: overloaded signatures
// written as one, parameters the stub marks required that PHP makes
// optional, or the reverse (lower case, `class::method` for methods).
var driftedStubs = map[string]bool{
	"array_keys": true, "array_walk": true, "array_walk_recursive": true, "assert_options": true,
	"closure::__invoke": true, "dateperiod::__construct": true, "dba_fetch": true,
	"get_class": true, "get_parent_class": true, "implode": true, "intlcal_set": true,
	"intlcalendar::fromdatetime": true, "intlcalendar::set": true,
	"intlgregcal_create_instance": true, "intlgregoriancalendar::__construct": true, "join": true,
	"ldap_exop": true, "mt_rand": true, "mysqli_stmt::__construct": true,
	"mysqli_stmt::bind_param": true, "mysqli_stmt::bind_result": true, "openssl_sign": true,
	"openssl_verify": true, "pg_clientencoding": true, "pg_errormessage": true, "pg_exec": true,
	"pg_execute": true, "pg_fetch_result": true, "pg_field_is_null": true,
	"pg_field_prtlen": true, "pg_fieldisnull": true, "pg_fieldprtlen": true, "pg_lo_create": true,
	"pg_lo_open": true, "pg_lo_unlink": true, "pg_locreate": true, "pg_loexport": true,
	"pg_loimport": true, "pg_loopen": true, "pg_lounlink": true, "pg_lowrite": true,
	"pg_prepare": true, "pg_query": true, "pg_query_params": true, "pg_result": true,
	"pg_set_error_verbosity": true, "pg_setclientencoding": true, "phar::setstub": true,
	"phardata::setstub": true, "rand": true, "reflectionclass::getstaticpropertyvalue": true,
	"reflectionproperty::setvalue": true, "session_set_cookie_params": true,
	"session_set_save_handler": true, "soapheader::__construct": true,
	"stream_context_set_option": true, "stream_filter_append": true,
	"stream_filter_prepend": true, "strtr": true,
}

// trustedBuiltin reports whether the stub declaration of callee (a function
// FQN or `Class::method`) in file states the parameter counts PHP enforces
// at version ver. Before PHP 8.0, builtins declared their parameters by
// hand and the stubs model them only roughly.
func trustedBuiltin(file, callee string, ver phpversion.Version) bool {
	callee = strings.ToLower(strings.TrimPrefix(callee, `\`))
	return ver.AtLeast(phpversion.PHP80) && semanticquery.BundledStub(file) && !driftedStubs[callee]
}
