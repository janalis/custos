<?php
class Form
{
    public function check(string $name, $id, int $n)
    {
        $a = <weak_warning descr="Compare with an empty string instead: '$name !== '''.">strlen($name) > 0</weak_warning>;
        $b = <weak_warning descr="Compare with an empty string instead: '(string)$id === '''.">!mb_strlen($id)</weak_warning>;
        $c = <weak_warning descr="Compare with an empty string instead: '(string)$n !== '''.">strlen($n) != 0</weak_warning>;
        $d = $a || <weak_warning descr="Compare with an empty string instead: '$name !== '''.">\strlen($name)</weak_warning>;
        do {} while (<weak_warning descr="Compare with an empty string instead: '$name === '''.">strlen($name) == 0</weak_warning>);
        $f = fn($s) => <weak_warning descr="Compare with an empty string instead: '(string)$s !== '''.">strlen($s) !== 0</weak_warning>;
        return [$a, $b, $c, $d, $f];
    }
}
