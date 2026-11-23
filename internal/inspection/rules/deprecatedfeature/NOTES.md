# DeprecatedFeature implementation notes

## Deprecation table

`data.go` lists PHP's own functions, methods and constants deprecated from
PHP 8.0 on, each with the version that deprecates it. `make deprecations`
writes it with `tools/gendeprecations`, from one PHP binary per release
from 8.0 to the newest (`DEPRECATION_PHP="8.0=/path/php ... 8.6=/path/php"`),
each with the extensions the table covers loaded:

1. The newest binary lists every deprecated name it declares: reflection's
   `isDeprecated()` for functions, methods, class constants and constants.
2. Each binary reports which of those names it deprecates: reflection for
   functions and methods, and reading the value for constants (some were
   deprecated in C, before reflection could tell).
3. A name dates from the oldest binary that deprecates it, which is why
   every release must be given.
4. `is_double`, `is_long`, `is_integer`, `doubleval`, `socket_set_timeout`
   and `mysqli_execute` are left to AliasFunctionsUsage.

The tool warns about a binary that lacks an extension declaring
deprecations, whose dates could come out late (8.0 and 8.1 lack `random`,
which split from `standard` in 8.2: harmless). Homebrew has no PHP 8.4
keg on every machine; one built from php-src (`git archive php-8.4.x`,
`./buildconf`, then `./configure --disable-all --enable-cli` with the
extensions the newest binary loads: mysqlnd, mysqli, pdo_mysql, pdo_pgsql,
pgsql, libxml, dom, xml, soap, curl, intl, fileinfo, sockets, session,
mbstring, openssl, shmop, zip, gd, odbc, filter, ctype, tokenizer) serves,
kept out of the repository (for instance under `.cache/php/8.4`).

Deprecations made before 8.0 all concern functions PHP 8.0 removed, which
UnavailableFunction reports. Rerun `make deprecations` for each PHP release,
and keep `hints` for the replacements PHP names.
