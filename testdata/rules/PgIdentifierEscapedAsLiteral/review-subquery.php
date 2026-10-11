<?php
function recent($c, $name) { $t = <warning descr="Escape PostgreSQL identifiers with the identifier API.">pg_escape_literal($c, $name)</warning>; return pg_query($c, "SELECT s.id FROM (SELECT id, max(seen) FROM $t GROUP BY id) s"); }
function known($c, $name) { $t = <warning descr="Escape PostgreSQL identifiers with the identifier API.">pg_escape_literal($c, $name)</warning>; return pg_query($c, "WITH k AS (SELECT id FROM stock WHERE id IN (1, 2)) SELECT k.id FROM k JOIN $t USING (id)"); }
