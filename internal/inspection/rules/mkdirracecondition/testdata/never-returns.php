<?php
function fail(string $m): void
{
    echo $m;
    exit(1);
}

function raise(string $m): void
{
    $f = function () { return 1; };
    throw new \RuntimeException($m);
}

function early(string $m): void
{
    if ($m === '') {
        return;
    }
    exit(1);
}

function talk(string $m): void
{
    echo $m;
}

function nothing(): void {}

function abort(): never
{
    exit(1);
}

function prepare(string $d): void
{
    mkdir($d) or fail('cannot');
    mkdir($d) or raise('cannot');
    mkdir($d) or abort();
    <error descr="Re-check with is_dir() after a failed mkdir: 'mkdir($d) || is_dir(...)'.">mkdir($d)</error> or early('cannot');
    <error descr="Re-check with is_dir() after a failed mkdir: 'mkdir($d) || is_dir(...)'.">mkdir($d)</error> or talk('cannot');
    <error descr="Re-check with is_dir() after a failed mkdir: 'mkdir($d) || is_dir(...)'.">mkdir($d)</error> or nothing();
    <error descr="Re-check with is_dir() after a failed mkdir: 'mkdir($d) || is_dir(...)'.">mkdir($d)</error> or undefined_fn();
    <error descr="Re-check with is_dir() after a failed mkdir: 'mkdir($d) || is_dir(...)'.">mkdir($d)</error> or $d();
}
