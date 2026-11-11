package argumentcountmismatch

import (
	"strings"

	"custos/internal/inspection/semanticquery"
	phpversion "custos/internal/php/version"
)

// driftedStubs lists the builtins of bundled directories whose stub
// declares other parameter counts than PHP does: overloaded signatures
// written as one, parameters the stub marks required that PHP makes
// optional, or the reverse (lower case, `class::method` for methods).
var driftedStubs = map[string]bool{
	"array_keys": true, "array_walk": true, "array_walk_recursive": true, "assert_options": true,
	"chop": true, "closure::__invoke": true, "collator::getsortkey": true,
	"collator::sortwithsortkeys": true, "collator_get_sort_key": true,
	"collator_sort_with_sort_keys": true, "datefmt_create": true, "datefmt_format": true,
	"dateperiod::__construct": true, "dba_fetch": true, "get_class": true, "get_parent_class": true,
	"implode": true, "intlcal_roll": true, "intlcal_set": true, "intlcalendar::fromdatetime": true,
	"intlcalendar::set": true, "intldateformatter::__construct": true,
	"intldateformatter::create": true, "intldateformatter::format": true,
	"intlgregcal_create_instance": true, "intlgregoriancalendar::__construct": true,
	"intltz_has_same_rules": true, "join": true, "ldap_exop": true, "ldap_exop_passwd": true,
	"ldap_free_result": true, "ldap_get_option": true, "ldap_parse_exop": true,
	"locale::filtermatches": true, "locale::getdisplaylanguage": true, "locale::getdisplayname": true,
	"locale::getdisplayregion": true, "locale::getdisplayscript": true,
	"locale::getdisplayvariant": true, "locale::lookup": true, "locale_filter_matches": true,
	"locale_get_display_language": true, "locale_get_display_name": true,
	"locale_get_display_region": true, "locale_get_display_script": true,
	"locale_get_display_variant": true, "locale_lookup": true, "ltrim": true, "mt_rand": true,
	"mysqli_stmt::__construct": true, "mysqli_stmt::bind_param": true,
	"mysqli_stmt::bind_result": true, "normalizer::isnormalized": true, "normalizer::normalize": true,
	"pg_clientencoding": true, "pg_errormessage": true, "pg_escape_bytea": true,
	"pg_escape_identifier": true, "pg_escape_literal": true, "pg_escape_string": true,
	"pg_exec": true, "pg_execute": true, "pg_fetch_result": true, "pg_field_is_null": true,
	"pg_field_prtlen": true, "pg_fieldisnull": true, "pg_fieldprtlen": true, "pg_lo_create": true,
	"pg_lo_open": true, "pg_lo_unlink": true, "pg_locreate": true, "pg_loexport": true,
	"pg_loimport": true, "pg_loopen": true, "pg_lounlink": true, "pg_lowrite": true,
	"pg_parameter_status": true, "pg_prepare": true, "pg_put_line": true, "pg_query": true,
	"pg_query_params": true, "pg_result": true, "pg_set_client_encoding": true,
	"pg_set_error_verbosity": true, "pg_setclientencoding": true, "phar::setstub": true,
	"phardata::setstub": true, "rand": true, "reflectionclass::getstaticpropertyvalue": true,
	"reflectionproperty::setvalue": true, "rtrim": true, "session_set_cookie_params": true,
	"session_set_save_handler": true, "soapheader::__construct": true, "socket_addrinfo_lookup": true,
	"stream_context_set_option": true, "stream_copy_to_stream": true, "stream_filter_append": true,
	"stream_filter_prepend": true, "stream_is_local": true, "stream_select": true,
	"stream_socket_pair": true, "strtr": true, "trim": true, "uconverter::getaliases": true,
	"uconverter::reasontext": true,
}

// trustedBuiltin reports whether the stub declaration of callee (a function
// FQN or `Class::method`) in file states the parameter counts PHP enforces
// at version ver. Before PHP 8.0, builtins declared their parameters by
// hand and the stubs model them only roughly.
func trustedBuiltin(file, callee string, ver phpversion.Version) bool {
	callee = strings.ToLower(strings.TrimPrefix(callee, `\`))
	return ver >= phpversion.PHP80 && semanticquery.BundledStub(file) && !driftedStubs[callee]
}
