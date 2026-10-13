<?php
class Kind { const List = 5; const LIST = 6; const Print = 'p'; const PRINT = 'P'; }
function kinds() {
    // Constants spelled like keywords are still case-sensitive names.
    $a = Kind::List ? Kind::LIST : 1;
    $b = Kind::Print ? Kind::PRINT : 1;
    $c = [Kind::List][0] ? [Kind::LIST][0] : 1;
    // Same spelling, and the case-insensitive `class` pseudo-constant.
    $d = Kind::List ?: 1;
    $e = Kind::class ?: 1;
    return [$a, $b, $c, $d, $e];
}
var_dump(kinds());
