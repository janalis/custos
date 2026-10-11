<?php
// Reads before the loop never see the value produced inside it.
function each($c, $names) {
    $t = null;
    foreach ($names as $name) {
        $run = function () { return 1; };
        $t = pg_escape_identifier($c, $name);
        pg_query($c, "SELECT 1 FROM $t LIMIT 0");
    }
}
