package astquery

import "strings"

// impureBuiltins lists built-in functions whose result differs between two
// evaluations or that change state (internal pointers, random generators,
// clocks, streams), lower-case.
var impureBuiltins = func() map[string]bool {
	m := map[string]bool{}
	for _, f := range strings.Fields(`next prev reset end each array_shift array_pop
		array_unshift array_push array_splice shuffle str_shuffle array_rand
		rand mt_rand random_int random_bytes lcg_value uniqid microtime hrtime time
		openssl_random_pseudo_bytes mcrypt_create_iv
		fgets fgetc fgetss fread fgetcsv fscanf fwrite fputs fputcsv readline
		stream_get_line stream_get_contents socket_read file_put_contents
		func_get_args`) {
		m[f] = true
	}
	return m
}()

// ImpureBuiltin reports whether the built-in function name, lower-case, may
// give another result when called again or changes state.
func ImpureBuiltin(name string) bool {
	return impureBuiltins[name]
}
