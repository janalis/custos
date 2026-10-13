<?php

function wrongTernaryArm($input)
{
    $clean = preg_replace('/\s+/u', ' ', $input);

    return $clean === null ? <warning descr="Handle regex replacement failure before using its result.">strtoupper($clean)</warning> : '';
}

function wrongOperand($input)
{
    $clean = preg_replace('/\s+/u', ' ', $input);

    return $clean !== null || <warning descr="Handle regex replacement failure before using its result.">strlen($clean)</warning> > 3;
}

function nonTerminatingGuard($input)
{
    $clean = preg_replace('/\s+/u', ' ', $input);
    if (is_null($clean)) {
        error_log('Replacement failed.');
    }

    return <warning descr="Handle regex replacement failure before using its result.">trim($clean)</warning>;
}

function replacedAfterGuard($input)
{
    $clean = preg_replace('/\s+/u', ' ', $input);
    if (!is_string($clean)) {
        return '';
    }
    $clean = preg_replace('/^-/u', '', $clean);

    return <warning descr="Handle regex replacement failure before using its result.">trim($clean)</warning>;
}

function guardOnOtherVariable($input, $other)
{
    $clean = preg_replace('/\s+/u', ' ', $input);
    if (is_string($other)) {
        return <warning descr="Handle regex replacement failure before using its result.">trim($clean)</warning>;
    }

    return '';
}
