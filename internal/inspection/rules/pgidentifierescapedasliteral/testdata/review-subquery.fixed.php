<?php
function recent($c, $name) { $t = pg_escape_identifier($c, $name); return pg_query($c, "SELECT s.id FROM (SELECT id, max(seen) FROM $t GROUP BY id) s"); }
function known($c, $name) { $t = pg_escape_identifier($c, $name); return pg_query($c, "WITH k AS (SELECT id FROM stock WHERE id IN (1, 2)) SELECT k.id FROM k JOIN $t USING (id)"); }
