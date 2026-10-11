<?php
final class Row
{
    public string $name = '';
    public int $id = 0;
}

function guards(?Row $o, $v, string $path, string $base)
{
    $file = \realpath($path);
    $fence = \realpath($base);

    // A filesystem call is no cheaper than in-memory string checks.
    if (false === $file || false === $fence || !\str_starts_with($file, \rtrim($fence, characters: '/') . '/') || !\is_file($file)) {}
    if (\strlen($path) > 3 && \file_exists($path)) {}
    if (\is_dir($path) && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$v</weak_warning>) {}

    // S5: an operand narrowing a variable stays ahead of its uses.
    if (!(null === $o || \strlen($o->name) === 0) && $o->id > 0) {}
    if (!(!$v instanceof \Countable || \count($v) === 0) && $v) {}
    if (!(!\is_string($v) || \strlen($v) === 0) && $v === 'x') {}
    if (!(\strlen($o->name) === 0 || $v == false) && $v > 1) {}
    if (\strlen($o->name) === \PHP_INT_MAX && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$o > 1</weak_warning>) {}
    if (!(null === $o || \strlen($o->name) === 0) && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$v > 0</weak_warning>) {}
    if (\in_array(1, [fn($o) => null === $o]) && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$o</weak_warning>) {} // a closure's own check
    if (!(\is_string($o->name) && \strlen($o->name) === 0) && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$o > 0</weak_warning>) {}
}

const LIMITS = ['small' => 1, 'large' => 9];

function guardedOperations(string $type, $k, array $keys, array $map, int $n, int $step, array $rows, string $s, bool $ok)
{
    // S6: an operation that can fail stays behind a check on its variables.
    if (\in_array($type, ['small', 'large'], true) && LIMITS[$type] > 5) {}
    if (\in_array($k, $keys, true) && $map[$k] > 1) {}
    if (\ctype_digit($k) && $map[(int) $k]) {}
    if (\strlen($s) === 3 && $map[$s[2]]) {}
    if (\abs($step) > 0 && $n % $step === 0) {}
    if (\count($rows) > 0 && $n / \count($rows) > 1) {}
    if (\abs($step) > 0 && \intdiv($n, $step) === 2) {}
    if (\abs($step) > 0 && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$n * $step === 0</weak_warning>) {}
    if (\abs($step) > 0 && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$step % $n === 0</weak_warning>) {}
    if (\abs($step) + \abs($n) > 0 && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">\intdiv($step, 2) === 1</weak_warning>) {}
    if (\strlen($s) > 3 && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$map[$k]</weak_warning>) {}
    if (\strlen($s) > 3 && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$n % 2 === 0</weak_warning>) {}
}
