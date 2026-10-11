<?php

function byReferenceCapture()
{
    $raw = '010';
    $read = function () use (&$raw) {
        $raw = '0x1F';
    };
    $read();

    return gmp_init($raw);
}

function incremented()
{
    $raw = '077';
    $raw++;

    return gmp_init($raw);
}

function copiedAfterIncrement()
{
    $raw = '077';
    $copy = $raw;
    $raw--;

    return gmp_init($copy);
}

function referencedArrayItem(array $rows)
{
    $raw = '010';
    $rows = [&$raw];
    fill($rows);

    return gmp_init($raw);
}

function extracted(array $input)
{
    $raw = '010';
    extract($input);

    return gmp_init($raw);
}

function parsedQuery($query)
{
    $raw = '010';
    parse_str($query);

    return gmp_init($raw);
}

function caught()
{
    $raw = '010';
    try {
        risky();
    } catch (Exception $raw) {
    } catch (Error) {
    }

    return gmp_init($raw);
}
