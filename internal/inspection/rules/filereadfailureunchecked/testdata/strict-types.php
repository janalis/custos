<?php declare(strict_types=1);
// Under strict types a false given for a string parameter throws a TypeError.
function hashText(string $text): string
{
    return sha1($text);
}

function run(string $path): string
{
    $code = file_get_contents($path);
    $hash = hashText($code);
    $trimmed = trim($code);
    return <warning descr="Handle file reading failure before using its result.">$code . $hash</warning> . $trimmed;
}
