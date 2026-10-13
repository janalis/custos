<?php

function falsy_exit($raw) {
    $parts = parse_url($raw);
    if (!$parts) {
        return null;
    }
    return $parts['host'];
}

function array_exit($raw) {
    $parts = parse_url($raw);
    if (!is_array($parts)) {
        throw new InvalidArgumentException('bad address');
    }
    return $parts['scheme'] . '://' . $parts['host'];
}

function array_branch($raw) {
    $parts = parse_url($raw);
    if (is_array($parts)) {
        return $parts['host'];
    }
    return null;
}

function key_probe($raw) {
    $parts = parse_url($raw);
    if (isset($parts['host'])) {
        return $parts['host'];
    }
    if (!empty($parts['path'])) {
        return $parts['path'];
    }
    return null;
}

function key_probe_exit($raw) {
    $parts = parse_url($raw);
    if (!isset($parts['scheme'], $parts['host'])) {
        return null;
    }
    return $parts['scheme'] . $parts['host'];
}

function conditional_expression($raw) {
    $parts = parse_url($raw);
    $host = $parts === false ? null : $parts['host'];
    $port = isset($parts['port']) ? $parts['port'] : 80;
    return $parts ? $host . ':' . $port . $parts['path'] : null;
}

function short_circuit($raw) {
    $parts = parse_url($raw);
    return ($parts !== false && $parts['host'] === 'example.test')
        || (!$parts || $parts['scheme'] === 'file');
}

function loose_and_empty($raw, $other) {
    $parts = parse_url($raw);
    if ($parts == false) {
        return null;
    }
    $second = parse_url($other);
    if (empty($second)) {
        return null;
    }
    return $parts['host'] . $second['host'];
}

function else_branches($raw, $strict) {
    $parts = parse_url($raw);
    if (!$parts) {
        return null;
    } elseif ($strict) {
        return $parts['host'];
    } else {
        return $parts['path'];
    }
}

function combined($raw) {
    $parts = parse_url($raw);
    if (!is_array($parts) || !isset($parts['host'])) {
        return null;
    }
    if (array_key_exists('port', $parts)) {
        return $parts['host'] . ':' . $parts['port'];
    }
    return $parts['host'];
}

function retried_with_prefix($raw, $prefix) {
    $attempts = 0;
    again:
    try {
        $parts = parse_url($raw);
        if (!is_array($parts)) {
            return null;
        }
        $host = $parts['host'];
        return $prefix . $host . ($parts[$prefix] ?? '');
    } catch (RuntimeException $failure) {
        if (++$attempts < 3) {
            goto again;
        }
        throw $failure;
    }
}
