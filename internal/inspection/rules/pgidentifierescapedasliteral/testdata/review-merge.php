<?php
// After the branch the comparison may receive either value, so the report
// stays and the rename is withheld.
function merged($c, $narrow, $label) {
    $v = pg_escape_literal($c, $label);
    if ($narrow) {
        $v = <warning descr="Escape PostgreSQL identifiers with the identifier API.">pg_escape_literal($c, 'stock')</warning>;
        pg_query($c, "SELECT 1 FROM $v LIMIT 0");
    }
    return pg_query($c, "SELECT 1 WHERE label = $v");
}
