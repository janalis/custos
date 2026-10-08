package stubs

import (
	"strings"

	"custos/internal/index"
	"custos/internal/phpver"
)

// pre80Failure lists builtins that returned false (or null) on invalid
// arguments before PHP 8.0, where 8.0 throws a ValueError instead, but whose
// phpstorm-stubs declaration gives only the 8.0 type (no
// LanguageLevelTypeAware map, no per-version declaration): `hash()` on 7.x
// is string|false. The value is the member added below 8.0.
var pre80Failure = map[string]string{
	"hash": "false", "hash_hmac": "false", "hash_pbkdf2": "false",
	"array_fill": "false", "chunk_split": "false", "wordwrap": "false",
	"substr_count": "false", "count_chars": "false", "str_word_count": "false",
	"array_chunk": "null", "array_rand": "null",
}

// addPre80Failures gives the functions of pre80Failure their pre-8.0
// return type (see index.VerType), unless the stubs already vary it.
func addPre80Failures(files []*index.FileSymbols) {
	for _, f := range files {
		for _, fn := range f.Functions {
			extra, ok := pre80Failure[strings.ToLower(fn.FQN)]
			if !ok || len(fn.RetVer) > 0 || fn.Return == "" {
				continue
			}
			fn.RetVer = []index.VerType{{Until: phpver.PHP80, Type: fn.Return + "|" + extra}}
		}
	}
}

// markBuiltin flags every declaration as a builtin and clears EmptyBody
// (stub bodies are placeholders).
func markBuiltin(files []*index.FileSymbols) {
	for _, f := range files {
		for _, fn := range f.Functions {
			fn.Builtin = true
		}
		for _, c := range f.Classes {
			for _, m := range c.Methods {
				m.Builtin, m.EmptyBody, m.StoresParams = true, false, false
			}
			for _, p := range c.Props {
				p.Builtin = true
			}
		}
	}
}

// missingProps lists the public properties extensions create at runtime on
// their objects that phpstorm-stubs does not declare (lower-case class
// FQN -> names). OAuthProvider (pecl/oauth) fills these from the request
// it checks; rules reading the stubs (MissingIssetImplementation) took
// them for undeclared.
var missingProps = map[string][]string{
	"oauthprovider": {"consumer_key", "consumer_secret", "signature", "signature_method", "token",
		"token_secret", "nonce", "timestamp", "version", "callback", "verifier"},
}

// addMissingProps declares the properties of missingProps on their stub
// classes (untyped, public), unless the stubs already do.
func addMissingProps(files []*index.FileSymbols) {
	for _, f := range files {
		for _, c := range f.Classes {
			for _, name := range missingProps[strings.ToLower(c.FQN)] {
				if c.Props[name] == nil {
					c.Props[name] = &index.Property{Name: name, Class: c.FQN, Visibility: index.Public}
				}
			}
		}
	}
}
