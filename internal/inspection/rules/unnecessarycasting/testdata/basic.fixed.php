<?php
$text  = 'abc';
$list  = [1];
$flag  = false;
$ratio = 1.5;
$count = 4;

return [
    $text,
    $list,
    $flag,
    $ratio,
    ($count * 3),
    ($count - 0.5),
    (int) ($count / 2),
    -2.25,
    (int) ($count * 0.5),
    (int) ($count + $missing),
    (string) $missing,
    (object) $list,
];

function strictArg(int $n, $loose) {
    return [$n, (int) $loose];
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
            (int) $this->hidden,
            (int) $this->shown,
            $this->size(),
            (int) $this->legacy(),
            (int) $other?->size(),
            (int) ($this->shown ?? 0),
        ];
    }
}

function joins(string $a, int $b, ?string $c, $d, \Stringable $e, string|false $f) {
    $a .= $b;
    $a .= (string) $c . (string) $d . (string) $e . (string) $f . (string) true;
    return 'n=' . $b . 2.5;
}

function stamps() {
    return [
        (microtime() * 10),
        (int) (microtime(true) * 10),
        (bool) end($GLOBALS),
        $_SERVER['HTTP_HOST'],
        $_SERVER['REQUEST_TIME'],
        (string) $_GET['q'],
        explode(',', 'a,b'),
        (string) str_replace('a', 'b', $unknown),
        str_replace('a', 'b', 'abc'),
    ];
}
