<?php
function lookup(array $matrix, $obj, $row, $col) {
    if (isset(<warning descr="Compute the concatenated key in a variable before using it.">$matrix['r' . $row]['c']</warning>)) {}
    if (isset(<warning descr="Compute the concatenated key in a variable before using it.">$matrix[$row]['c' . $col][0]</warning>)) {}
    if (!isset(<warning descr="Compute the concatenated key in a variable before using it.">$matrix['r' . $row][$col]</warning>)) {}
    if (isset($obj->{'p' . $row}['c'])) {}
    if (isset($matrix[('r' . $row)]['c'])) {}
    $kept = isset($matrix['r' . $row]['c']);
    return $kept;
}
