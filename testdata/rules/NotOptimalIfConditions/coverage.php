<?php
class Repo {
    public function take(&$v) { return 1; }
    public static function grab(&$v) { return 1; }
}
interface Shape {}
// costs of the remaining expression kinds
if (strlen("a{$x}") && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$y</weak_warning>) {}
if (strlen($s) && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$i++</weak_warning>) {}
if (strlen($s) && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">[1, 'k' => $v]</weak_warning>) {}
if (strlen($s) && (<weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$c ? $d : $e</weak_warning>)) {}
if (($c ? strlen($s) : $e) && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$d</weak_warning>) {}
if (($c ? $e : strlen($s)) && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$d</weak_warning>) {}
// S1: destructuring targets
if (([$first] = explode(',', $csv)) && $first) {}
if ((list(, $second) = explode(',', $csv)) && $second) {}
// S2: nested array accesses
if (strlen($k) && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$map['a']['b']</weak_warning>) {}
if (strlen($map['a']) && $map['a']['b']) {}
if (strlen($other['x']) && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$map['a']['b']</weak_warning>) {}
// S3: guard hit inside a larger operand
if (isset($rows[md5($k)]) && $rows > 0) {}
// S1 with argument variants
if (sscanf($line, '%d %d', $w, $h) && $h) {}
if (strlen(string: $s) && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$s</weak_warning>) {}
if ($o->$m($v) && $v) {}
$repo = new Repo();
if ($repo->take($v) && $v) {}
if (Repo::grab($v) && $v) {}
if ($cls::grab($v) && $v) {}
if (Repo::$fn($v) && $v) {}
// D3 / D4 corner cases
if ($a instanceof Shape && <weak_warning descr="Equality check on a value also tested with instanceof; verify the logic.">$a === -1</weak_warning>) {}
if ($a instanceof Shape && $a > 1) {}
if ($a instanceof $cls || $a instanceof Shape) {}
