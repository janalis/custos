<?php
function totals($x)
{
    // A comment the fix would drop: no fix.
    <weak_warning descr="Use the compound form '$x += 1'.">$x = $x /* c */ + 1</weak_warning>;
    return $x;
}
