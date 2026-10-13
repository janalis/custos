<?php

function wrongBound($in)
{
    $raw = fread($in, 4);
    if (strlen($raw) < 3) {
        return null;
    }
    return <warning descr="Read the complete record before decoding it.">unpack('Nsize', $raw)</warning>;
}

function onlyFailureExcluded($in)
{
    $raw = fread($in, 4);
    if ($raw === false || $raw === '') {
        return null;
    }
    return <warning descr="Read the complete record before decoding it.">unpack('Nsize', $raw)</warning>;
}

function eitherOperandMayHold($in, $force)
{
    $raw = fread($in, 4);
    if (strlen($raw) === 4 || $force) {
        return <warning descr="Read the complete record before decoding it.">unpack('Nsize', $raw)</warning>;
    }
    return null;
}

function guardDoesNotLeave($in)
{
    $raw = fread($in, 4);
    if (strlen($raw) !== 4) {
        error_log('truncated');
    }
    return <warning descr="Read the complete record before decoding it.">unpack('Nsize', $raw)</warning>;
}

function otherValueChecked($in, $other)
{
    $raw = fread($in, 4);
    if (strlen($other) !== 4) {
        return null;
    }
    return <warning descr="Read the complete record before decoding it.">unpack('Nsize', $raw)</warning>;
}

function characterCount($in)
{
    $raw = fread($in, 4);
    if (mb_strlen($raw) !== 4) {
        return null;
    }
    return <warning descr="Read the complete record before decoding it.">unpack('Nsize', $raw)</warning>;
}
