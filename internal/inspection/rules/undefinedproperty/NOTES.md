# UndefinedProperty implementation notes

## Shared check

The rule shares the receiver resolution, hierarchy walk, subclass walk
and guard search of UndefinedMethod (see its notes). Its own parts:

- `read` climbs through parentheses and element fetches to tell a read
  from a write; arguments use `semanticquery.ArgumentMayWrite`, so an
  argument to an unresolved callee counts as a write;
- `checked` climbs through element and property fetches to find `isset`,
  `empty` or the left of `??`;
- `written` lists, once per file through `analysis.Context.Shared`, the
  instance property names the file writes (by `read`'s terms): a read of
  one of these is never reported;
- the member test accepts a property, `__get`, `__isset`, `stdClass` and
  `#[AllowDynamicProperties]`.

## Corpus triage

The corpora (guzzle, composer, laravel) give no finding; before the
subclass walk, laravel gave 4 (an interface or base class whose
implementation declares the property).
