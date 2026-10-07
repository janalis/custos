<?php
$text  = 'abc';
$list  = [1];
$flag  = false;
$ratio = 1.5;
$count = 4;

return [
    <weak_warning descr="Operand already has the target type; remove the cast.">(string)</weak_warning> $text,
    <weak_warning descr="Operand already has the target type; remove the cast.">(array)</weak_warning> $list,
    <weak_warning descr="Operand already has the target type; remove the cast.">(bool)</weak_warning> $flag,
    <weak_warning descr="Operand already has the target type; remove the cast.">(double)</weak_warning> $ratio,
    <weak_warning descr="Operand already has the target type; remove the cast.">(integer)</weak_warning> ($count * 3),
    <weak_warning descr="Operand already has the target type; remove the cast.">(float)</weak_warning> ($count - 0.5),
    (int) ($count / 2),
    <weak_warning descr="Operand already has the target type; remove the cast.">(float)</weak_warning>-2.25,
    (int) ($count * 0.5),
    (int) ($count + $missing),
    (string) $missing,
    (object) $list,
];

function strictArg(int $n, $loose) {
    return [<weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning> $n, (int) $loose];
}

class Meter {
    /** @var int */
    private $hidden = 0;
    /** @var int */
    public $shown;
    public function size(): int { return 1; }
    /** @return int */
    public function legacy() { return 1; }

    public function read(?Meter $other) {
        return [
            <weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning> $this->hidden,
            (int) $this->shown,
            <weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning> $this->size(),
            (int) $this->legacy(),
            (int) $other?->size(),
            (int) ($this->shown ?? 0),
        ];
    }
}

function joins($a, $b) {
    $a .= <weak_warning descr="Concatenation converts to string anyway; remove the cast.">(string)</weak_warning> $b;
    return 'n=' . <weak_warning descr="Concatenation converts to string anyway; remove the cast.">(string)</weak_warning>$b;
}

function stamps() {
    return [
        <weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning> (microtime() * 10),
        (int) (microtime(true) * 10),
        (bool) end($GLOBALS),
        <weak_warning descr="Operand already has the target type; remove the cast.">(string)</weak_warning> $_SERVER['HTTP_HOST'],
        <weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning> $_SERVER['REQUEST_TIME'],
        (string) $_GET['q'],
        <weak_warning descr="Operand already has the target type; remove the cast.">(array)</weak_warning> explode(',', 'a,b'),
        (string) str_replace('a', 'b', $unknown),
        <weak_warning descr="Operand already has the target type; remove the cast.">(string)</weak_warning> str_replace('a', 'b', 'abc'),
    ];
}
