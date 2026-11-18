<?php
function good($path) { $h = fopen($path, 'rb'); if ($h === false) { return; } fread($h, 12); fclose($h); }

function falsyThrow(string $path): string
{
    $h = fopen($path, 'rb');
    if (!$h) {
        throw new RuntimeException('cannot open');
    }
    $head = fread($h, 4);
    fclose($h);
    return $head;
}

function assignedInCondition(string $path): ?string
{
    if ($h = fopen($path, 'rb')) {
        $line = fgets($h);
        fclose($h);
        return $line;
    }
    return null;
}

function comparedAssignment(string $path): void
{
    if (($h = fopen($path, 'ab')) !== false) {
        fwrite($h, 'x');
    }
}

function truthyBranch(string $path): string
{
    $h = fopen($path, 'rb');
    $out = '';
    if ($h) {
        while (($line = fgets($h)) !== false) {
            $out .= $line;
        }
        fclose($h);
    }
    return $out;
}

function resourceGuard(string $path): string
{
    $h = fopen($path, 'rb');
    if (!is_resource($h)) {
        return '';
    }
    return fread($h, 10);
}

function emptyGuard(string $path): void
{
    $h = fopen($path, 'rb');
    if (empty($h) || false == $h) {
        return;
    }
    rewind($h);
}

function orDie(string $path): void
{
    $h = fopen($path, 'rb') or die('cannot open');
    fread($h, 1);
}

function orThrow(string $path): void
{
    $h = fopen($path, 'rb');
    $h !== false || throw new RuntimeException('cannot open');
    fseek($h, 2);
    $g = fopen($path, 'rb');
    $g ?: throw new RuntimeException('cannot open');
    ftell($g);
}

function expressions(string $path)
{
    $h = fopen($path, 'cb');
    $written = $h && fwrite($h, 'x');
    $first = $h === false ? null : fread($h, 1);
    return $h ? [$written, $first, fflush($h)] : null;
}

function loops(array $paths, string $path): void
{
    foreach ($paths as $p) {
        $h = fopen($p, 'rb');
        if (!$h) {
            continue;
        }
        fread($h, 1);
    }
    $h = fopen($path, 'rb');
    while ($h && !feof($h)) {
        fgets($h);
    }
}

function elseBranch(string $path): void
{
    $h = fopen($path, 'cb');
    if ($h === false) {
        echo 'failed';
    } elseif ($path !== '') {
        ftruncate($h, 0);
    } else {
        fwrite($h, 'x');
    }
}
