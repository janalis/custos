<?php

function validated(string $path): array
{
    $json = file_get_contents($path);
    if (!json_validate($json)) {
        throw new \InvalidArgumentException('Invalid JSON.');
    }

    return json_decode($json, true) + ['size' => strlen($json)];
}

function validatedInBranch(string $path): int
{
    $json = file_get_contents($path);
    if (json_validate($json)) {
        return strlen($json);
    }

    return <warning descr="Handle file reading failure before using its result.">strlen($json)</warning>;
}
