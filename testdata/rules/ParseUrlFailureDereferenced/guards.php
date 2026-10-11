<?php

function unrelated_condition($raw, $verbose) {
    $parts = parse_url($raw);
    if ($verbose) {
        echo <warning descr="Reject URL parsing failure before indexing.">$parts['host']</warning>;
    }
}

function failure_branch($raw) {
    $parts = parse_url($raw);
    if (!$parts) {
        return <warning descr="Reject URL parsing failure before indexing.">$parts['host']</warning>;
    }
    return null;
}

function whole_variable_probe($raw) {
    $parts = parse_url($raw);
    if (isset($parts)) {
        return <warning descr="Reject URL parsing failure before indexing.">$parts['host']</warning>;
    }
    return null;
}

function stale_guard($raw, $next) {
    $parts = parse_url($raw);
    if (!$parts) {
        return null;
    }
    $parts = parse_url($next);
    return <warning descr="Reject URL parsing failure before indexing.">$parts['host']</warning>;
}

function either_operand($raw, $fallback) {
    $parts = parse_url($raw);
    if ($parts || $fallback) {
        return <warning descr="Reject URL parsing failure before indexing.">$parts['host']</warning>;
    }
    return null;
}

function assertion_only($raw) {
    $parts = parse_url($raw);
    assert($parts !== false);
    return <warning descr="Reject URL parsing failure before indexing.">$parts['host']</warning>;
}
