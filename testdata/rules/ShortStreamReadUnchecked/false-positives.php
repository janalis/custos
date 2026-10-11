<?php

function lowerBound($in)
{
    $raw = fread($in, 4);
    if (strlen($raw) < 4) {
        throw new RuntimeException('truncated');
    }
    return unpack('Nsize', $raw);
}

function failureOrLength($in)
{
    $raw = fread($in, 4);
    if ($raw === false || strlen($raw) !== 4) {
        throw new RuntimeException('truncated');
    }
    return unpack('Nsize', $raw);
}

function literalFirst($in)
{
    $raw = fread($in, 4);
    if (4 !== strlen($raw)) {
        return null;
    }
    return unpack('Vsize', $raw);
}

function looseComparison($in)
{
    $raw = fread($in, 2);
    if (strlen($raw) != 2) {
        return null;
    }
    $position = ftell($in);
    return array($position, unpack('nsize', $raw));
}

function conjunction($in)
{
    $raw = fread($in, 2);
    if (strlen($raw) === 2 && $raw !== "\0\0") {
        return unpack('vsize', $raw);
    }
    return null;
}

function singleByte($in)
{
    $raw = fread($in, 1);
    if ($raw === false || $raw === '') {
        return null;
    }
    return unpack('Cflag', $raw);
}

function singleByteTruthy($in)
{
    $raw = fread($in, 1);
    if (!$raw) {
        return null;
    }
    return unpack('cflag', $raw);
}

function byteCount($in)
{
    $raw = fread($in, 4);
    if (mb_strlen($raw, '8bit') !== 4) {
        return null;
    }
    return unpack('Vsize', $raw);
}

function lastOffset($in)
{
    $raw = fread($in, 4);
    if (!isset($raw[3])) {
        return null;
    }
    return unpack('Vsize', $raw);
}

function alternatives($in)
{
    $raw = fread($in, 4);
    if ($raw === false) {
        return null;
    } elseif (strlen($raw) >= 4) {
        return unpack('Nsize', $raw);
    }
    if (strlen($raw) > 3) {
        return null;
    } else {
        return strlen($raw) === 4 ? unpack('Nsize', $raw) : null;
    }
}

function shortCircuit($in)
{
    $raw = fread($in, 4);
    return strlen($raw) === 4 && unpack('Nsize', $raw);
}
