<?php
class Lock
{
    public function __destruct()
    {
        echo 'released';
    }
}

function buffer(string $s): ?FFI\CData { return null; }

function encode(array $headers, Lock $lock)
{
    // To keep memory buffers alive while C code reads them.
    $buffers = [];
    $held = [];
    $plain = [];
    foreach ($headers as $name) {
        $buf = buffer($name);
        $buffers[] = $buf;
        $held[$name] = $lock;
        <weak_warning descr="Value is only written here and never read; the write is lost.">$plain[]</weak_warning> = $name;
    }
    $mixed = [];
    <weak_warning descr="Value is only written here and never read; the write is lost.">$mixed[]</weak_warning> = buffer('x');
    <weak_warning descr="Value is only written here and never read; the write is lost.">$mixed[]</weak_warning> = 'y';
    $whole = [];
    $whole = buffer('z');
    return count($headers);
}
