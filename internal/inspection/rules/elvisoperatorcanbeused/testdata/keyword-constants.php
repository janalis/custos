<?php
class Kind { const List = 5; const LIST = 6; const Print = 'p'; const PRINT = 'P'; }
function kinds() {
    // Constants spelled like keywords are still case-sensitive names.
    $a = Kind::List ? Kind::LIST : 1;
    $b = Kind::Print ? Kind::PRINT : 1;
    $c = [Kind::List][0] ? [Kind::LIST][0] : 1;
    // Same spelling, and the case-insensitive `class` pseudo-constant.
    $d = <weak_warning descr="Use the short ternary: 'Kind::List ?: 1'.">Kind::List ? kind::List : 1</weak_warning>;
    $e = <weak_warning descr="Use the short ternary: 'Kind::class ?: 1'.">Kind::class ? Kind::CLASS : 1</weak_warning>;
    return [$a, $b, $c, $d, $e];
}
var_dump(kinds());
