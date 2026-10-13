<?php

function strictTernary($input)
{
    $clean = preg_replace('/\s+/u', ' ', $input);

    return $clean === null ? '' : strtoupper($clean);
}

function nullPredicateReturn($input)
{
    $clean = preg_replace('/\s+/u', ' ', $input);
    if (is_null($clean)) {
        return '';
    }

    return trim($clean);
}

function stringPredicateBranch($input)
{
    $clean = preg_replace_callback('/\d+/u', 'strrev', $input);
    if (is_string($clean)) {
        return strtoupper($clean);
    }

    return '';
}

function negatedStringPredicate($input)
{
    $clean = preg_replace('/\s+/u', ' ', $input);
    if (!is_string($clean)) {
        throw new RuntimeException('Invalid subject.');
    }
    error_log('normalised');

    return strtolower($clean);
}

function shortCircuitAnd($input)
{
    $clean = preg_replace('/\s+/u', ' ', $input);

    return $clean !== null && strlen($clean) > 3;
}

function shortCircuitOr($input)
{
    $clean = preg_replace('/\s+/u', ' ', $input);
    if ($clean === null || strlen($clean) === 0) {
        return '';
    }

    return $clean;
}

function issetTernary($input)
{
    $clean = preg_replace('/\s+/u', ' ', $input);

    return isset($clean) ? strtoupper($clean) : '';
}

function elseifBranch($input, $raw)
{
    $clean = preg_replace('/\s+/u', ' ', $input);
    if ($raw) {
        return $input;
    } elseif (is_string($clean)) {
        return trim($clean);
    }

    return '';
}

function assignmentInCondition($input)
{
    if (null !== $clean = preg_replace('/\s+/u', ' ', $input)) {
        return trim($clean);
    }

    return '';
}
