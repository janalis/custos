<?php
// Reads before the loop never see the value produced inside it.
function each($c, $names) {
    $t = null;
    foreach ($names as $name) {
        $run = function () { return 1; };
        $t = <warning descr="Escape PostgreSQL identifiers with the identifier API.">pg_escape_literal($c, $name)</warning>;
        pg_query($c, "SELECT 1 FROM $t LIMIT 0");
    }
}
