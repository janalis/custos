<?php
function earlyExit(string $dir): void
{
    if (is_dir($dir)) {
        return;
    }
    @mkdir($dir, 0777, true);
}

function elseArm(string $dir): void
{
    if (is_dir($dir)) {
        echo 'present';
    } else {
        @mkdir($dir);
    }
}

function elseifArm(string $dir, bool $skip): void
{
    if ($skip) {
        echo 'skipped';
    } elseif (!is_dir($dir)) {
        @mkdir($dir);
    }
}

function shortCircuit(string $dir): void
{
    is_dir($dir) || @mkdir($dir);
}

function noExit(string $dir): void
{
    if (is_dir($dir)) {
        echo 'present';
    }
    if (!@mkdir($dir) && !is_dir($dir)) { throw new \RuntimeException(sprintf('Directory "%s" was not created', $dir)); }
}

function otherDirectory(string $dir, string $log): void
{
    if (is_dir($log)) {
        return;
    }
    if (!@mkdir($dir) && !is_dir($dir)) { throw new \RuntimeException(sprintf('Directory "%s" was not created', $dir)); }
}

function unknownArgument(string $dir, array $parts): void
{
    $check = is_dir(...);
    if (is_dir(...$parts)) {
        return;
    }
    if (!@mkdir($dir) && !is_dir($dir)) { throw new \RuntimeException(sprintf('Directory "%s" was not created', $dir)); }
}
