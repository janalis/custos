<?php

function exclusive_or($raw, $strict) {
    if ($strict xor ($parts = parse_url($raw)) === false) {
        return null;
    }
    return <warning descr="Reject URL parsing failure before indexing.">$parts['host']</warning>;
}

function compared_conditions($raw, $strict) {
    if ($strict == (($parts = parse_url($raw)) === false)) {
        return null;
    }
    return <warning descr="Reject URL parsing failure before indexing.">$parts['host']</warning>;
}

function logged_only($raw) {
    if (($parts = parse_url($raw)) === false) {
        error_log('bad url');
    }
    return <warning descr="Reject URL parsing failure before indexing.">$parts['host']</warning>;
}

function exits_on_failure($raw) {
    if (($parts = parse_url($raw)) === false) {
        return null;
    }
    return $parts['host'];
}

function truthy_branch($raw) {
    if ($parts = parse_url($raw)) {
        return $parts['host'];
    }
    return null;
}
