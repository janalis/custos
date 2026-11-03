<?php
// A failed read contains no text: the test takes the branch of an empty file.
function scan(array $paths): void
{
    foreach ($paths as $path) {
        $code = file_get_contents($path);
        if (!str_contains($code, '@marker')) {
            continue;
        }
        echo $code;
    }
}

function head(string $path): string
{
    $code = file_get_contents($path);
    if (str_starts_with(haystack: $code, needle: ('<?php'))) {
        return 'php';
    }
    return str_ends_with(file_get_contents($path), "\n") ? 'eol' : '';
}

function stored(string $path): bool
{
    return $found = <warning descr="Handle file reading failure before using its result.">str_contains(file_get_contents($path), '@marker')</warning>;
}

function empty_needle(string $path): bool
{
    return !<warning descr="Handle file reading failure before using its result.">str_contains(file_get_contents($path), '')</warning>;
}

function variable_needle(string $path, string $marker): bool
{
    return !<warning descr="Handle file reading failure before using its result.">str_contains(file_get_contents($path), $marker)</warning>;
}

function needle(string $path): bool
{
    return !<warning descr="Handle file reading failure before using its result.">str_contains('@marker', file_get_contents($path))</warning>;
}

function other(string $path): bool
{
    return !<warning descr="Handle file reading failure before using its result.">strstr(file_get_contents($path), '@marker')</warning>;
}
