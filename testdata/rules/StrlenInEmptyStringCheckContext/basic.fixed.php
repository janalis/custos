<?php
class Form
{
    public function check(string $name, $id, int $n)
    {
        $a = $name !== '';
        $b = (string)$id === '';
        $c = (string)$n !== '';
        $d = $a || $name !== '';
        do {} while ($name === '');
        $f = fn($s) => (string)$s !== '';
        return [$a, $b, $c, $d, $f];
    }
}
