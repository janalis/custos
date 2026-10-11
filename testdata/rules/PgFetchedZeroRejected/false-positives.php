<?php
// A first-class callable creates a Closure and fetches nothing.
function callable_assigned() { if ($fetch = pg_fetch_result(...)) { return $fetch; } }
function callable_bare() { if (pg_fetch_result(...)) { echo 'closure'; } }
// Compound assignments test the combined value, never the fetched one.
function concat($r) { $v = ''; if ($v .= pg_fetch_result($r, 0, 0)) { echo $v; } }
function coalesce($r, $v) { if ($v ??= pg_fetch_result($r, 0, 0)) { echo $v; } }
// Targets that may coerce or reshape the fetched value.
function property($r, $o) { if ($o->p = pg_fetch_result($r, 0, 0)) { echo $o->p; } }
function destructured($r) { if ([$v] = pg_fetch_result($r, 0, 0)) { echo $v; } }
function dynamic($r, $name) { if ($$name = pg_fetch_result($r, 0, 0)) { echo 'set'; } }
function element($r, array $a) { if ($a['v'] = pg_fetch_result($r, 0, 0)) { echo $a['v']; } }
// Negated and combined conditions are deliberate.
function negated($r) { if (!pg_fetch_result($r, 0, 0)) { echo 'none'; } }
function combined($r, $ok) { if ($ok && pg_fetch_result($r, 0, 0)) { echo 'both'; } }
function statement($r) { pg_fetch_result($r, 0, 0); }
