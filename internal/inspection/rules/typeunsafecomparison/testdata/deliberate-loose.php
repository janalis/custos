<?php
class Name { public function __toString(): string { return 'n'; } }
class Identifier { public function __toString(): string { return 'i'; } }
class ComplexType {}

/** @return string|false */
function doc() { return false; }

function compare(Name|Identifier|ComplexType $type, ComplexType $complex, $v, string $s, int $n, ?string $opt)
{
    // Members that convert to a string make the comparison meaningful.
    $a = $type == 'void';
    $b = <error descr="\ComplexType has no __toString(), so it cannot be compared to a string.">$complex == 'void'</error>;
    // Emptiness tests on a value that may be false or null.
    $c = doc() != '';
    $d = false == doc();
    $e = doc() == 0;
    $f = $opt == null;
    $g = '0' != $opt;
    $h = <warning descr="Use '!==' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">doc() != 'x'</warning>;
    $i = <weak_warning descr="Prefer '===' to avoid implicit type juggling.">$s == ''</weak_warning>;
    $j = $v == false;
    $k = <weak_warning descr="Prefer '===' to avoid implicit type juggling.">$n == 1</weak_warning>;
    $l = <weak_warning descr="Prefer '===' to avoid implicit type juggling.">$opt == \PHP_EOL</weak_warning>;
    // A value compared with its own conversion tests what it keeps.
    $m = (int) $v == $v;
    $o = $v != (float) ($v);
    $p = "$n" != $n;
    $q = <weak_warning descr="Prefer '===' to avoid implicit type juggling.">(int) $v == $s</weak_warning>;
    $r = <weak_warning descr="Prefer '===' to avoid implicit type juggling.">(bool) $v == $v</weak_warning>;
    $t = <weak_warning descr="Prefer '===' to avoid implicit type juggling.">"$n$n" == $n</weak_warning>;
    return [$a, $b, $c, $d, $e, $f, $g, $h, $i, $j, $k, $l, $m, $o, $p, $q, $r, $t];
}

function storedConversion(float $value, int $n, bool $round)
{
    // The conversion stored in a local first still tests what it keeps.
    $repr = (string) $value;
    $a = $repr == $value;
    $text = 'n';
    if ($round) {
        $text = (string) $n;
    }
    $b = <weak_warning descr="Prefer '===' to avoid implicit type juggling.">$text == $n</weak_warning>;
    $c = <weak_warning descr="Prefer '===' to avoid implicit type juggling.">$fresh == $n</weak_warning>;
    $d = <weak_warning descr="Prefer '===' to avoid implicit type juggling.">$repr == $n</weak_warning>;
    return [$a, $b, $c, $d];
}

function truthiness(int|bool $flag, mixed $m, int $n, $unknown)
{
    // A flag stored as 1 or TRUE, a mixed value tested for emptiness.
    $a = $flag == true;
    $b = $m == null;
    $c = $m != '';
    $d = TRUE == $m;
    $e = <weak_warning descr="Prefer '===' to avoid implicit type juggling.">$n == true</weak_warning>;
    $f = $unknown == true;
    $g = <weak_warning descr="Prefer '===' to avoid implicit type juggling.">$unknown == null</weak_warning>;
    return [$a, $b, $c, $d, $e, $f, $g];
}
