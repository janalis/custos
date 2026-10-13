<?php
// The comparison at the top of the body reads the escaped value on the next
// iteration, so the report stays and the rename is withheld.
function carried($c, $label) {
    $v = pg_escape_literal($c, $label);
    for ($i = 0; $i < 2; $i++) {
        pg_query($c, "SELECT 1 WHERE label = $v");
        $v = <warning descr="Escape PostgreSQL identifiers with the identifier API.">pg_escape_literal($c, 'stock')</warning>;
        pg_query($c, "SELECT 1 FROM $v LIMIT 0");
    }
}
