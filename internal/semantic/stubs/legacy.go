package stubs

import (
	"strings"

	phpversion "github.com/janalis/custos/internal/php/version"
	"github.com/janalis/custos/internal/semantic/index"
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
			fn.RetVer = []index.VerType{{Until: phpversion.PHP80, Type: fn.Return + "|" + extra}}
		}
	}
}

// bundled lists the extensions that ship with the latest PHP, named as
// Extensions names them: the PHP versions the stubs give their
// declarations are PHP's, and no other build of them declares what PHP
// removed. Others are PECL or vendor extensions, whose stubs may give the
// extension's own versions (the New Relic agent's 8.4), or extensions PHP
// moved to PECL (imap, pspell, mysql, mcrypt…).
var bundled = map[string]bool{
	"bcmath": true, "bz2": true, "calendar": true, "com_dotnet": true, "core": true, "ctype": true,
	"curl": true, "date": true, "dba": true, "dom": true, "enchant": true, "exif": true, "ffi": true,
	"fileinfo": true, "filter": true, "fpm": true, "ftp": true, "gd": true, "gettext": true, "gmp": true,
	"hash": true, "iconv": true, "intl": true, "json": true, "ldap": true, "libxml": true,
	"mbstring": true, "mysqli": true, "odbc": true, "openssl": true, "pcntl": true, "pcre": true,
	"pdo": true, "pgsql": true, "phar": true, "posix": true, "random": true,
	"readline": true, "reflection": true, "session": true, "shmop": true, "simplexml": true,
	"snmp": true, "soap": true, "sockets": true, "sodium": true, "spl": true, "sqlite3": true,
	"standard": true, "sysvmsg": true, "sysvsem": true, "sysvshm": true, "tidy": true,
	"tokenizer": true, "uri": true, "xml": true, "xmlreader": true, "xmlwriter": true, "xsl": true,
	"zend-opcache": true, "zip": true, "zlib": true,
}

// extensionOf names the extension of a stub file as Extensions does ("" for
// a file outside an extension directory).
func extensionOf(path string) string {
	dir, _, ok := strings.Cut(strings.TrimPrefix(path, "stubs/"), "/")
	if !ok {
		return ""
	}
	return strings.ReplaceAll(strings.ToLower(dir), " ", "-")
}

// markBuiltin flags every declaration as a builtin and clears EmptyBody
// (stub bodies are placeholders).
func markBuiltin(files []*index.FileSymbols) {
	for _, f := range files {
		ext := extensionOf(f.Path)
		for _, c := range f.Constants {
			c.Builtin = true
		}
		for _, fn := range f.Functions {
			fn.Builtin, fn.Ext, fn.Bundled = true, ext, bundled[ext]
		}
		for _, c := range f.Classes {
			for _, m := range c.Methods {
				m.Builtin, m.EmptyBody, m.StoresParams = true, false, false
				for _, variant := range m.Variants {
					variant.Builtin, variant.EmptyBody, variant.StoresParams = true, false, false
				}
			}
			for _, p := range c.Props {
				p.Builtin = true
			}
		}
	}
}

// missingProps lists the public properties extensions create at runtime on
// their objects that phpstorm-stubs does not declare (lower-case class
// FQN -> names). See rules/missingissetimplementation/NOTES.md.
var missingProps = map[string][]string{
	"oauthprovider": {
		"consumer_key", "consumer_secret", "signature", "signature_method", "token",
		"token_secret", "nonce", "timestamp", "version", "callback", "verifier",
	},
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
