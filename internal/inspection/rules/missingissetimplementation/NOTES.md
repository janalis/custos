# MissingIssetImplementation implementation notes

## Stub gaps

Some extensions create public properties at runtime that phpstorm-stubs
does not declare. `OAuthProvider` (pecl/oauth) fills `consumer_key`,
`token`, `nonce` and similar properties from the request it checks; read
from the bare stubs, `isset($provider->token)` looked like a check of an
undeclared property. `internal/semantic/stubs/legacy.go` (`missingProps`)
declares these properties on the stub classes so this rule sees them.
