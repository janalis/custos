<?php
function totals($x)
{
    // A comment the fix would drop: no fix.
    $x = $x /* c */ + 1;
    return $x;
}
