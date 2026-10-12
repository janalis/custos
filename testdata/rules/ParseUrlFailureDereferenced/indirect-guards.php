<?php
function fail(string $message): never
{
    throw new RuntimeException($message);
}

function report(string $message): void
{
    error_log($message);
}

// The read is only reached when the parse succeeded.
function leavesOnFailure($url)
{
    $parts = parse_url($url);
    $parts !== false or exit(1);
    echo $parts['host'];
}

function throwsOnFailure($url)
{
    $parts = parse_url($url) or throw new InvalidArgumentException($url);
    return $parts['host'];
}

function neverReturns($url)
{
    $parts = parse_url($url);
    if ($parts === false) {
        fail('bad url');
    }
    return $parts['host'];
}

function storedTest($url)
{
    $parts = parse_url($url);
    $valid = $parts !== false;
    if (!$valid) {
        return null;
    }
    return $parts['host'];
}

function arrayOnlyCall($url)
{
    $parts = parse_url($url);
    if (count($parts) > 1) {
        return $parts['host'];
    }
    return null;
}

// The evidence does not hold.
function returnsOnFailure($url)
{
    $parts = parse_url($url);
    if ($parts === false) {
        report('bad url');
    }
    return <warning descr="Reject URL parsing failure before indexing.">$parts['host']</warning>;
}

function leavesOnSuccess($url)
{
    $parts = parse_url($url);
    $parts === false or exit(1);
    echo <warning descr="Reject URL parsing failure before indexing.">$parts['host']</warning>;
}

function storedTestOfEarlierResult($url, $other)
{
    $parts = parse_url($url);
    $valid = $parts !== false;
    $parts = parse_url($other);
    if (!$valid) {
        return null;
    }
    return <warning descr="Reject URL parsing failure before indexing.">$parts['host']</warning>;
}

function storedTestOverwritten($url, $flag)
{
    $parts = parse_url($url);
    $valid = $parts !== false;
    $valid = $flag;
    if (!$valid) {
        return null;
    }
    return <warning descr="Reject URL parsing failure before indexing.">$parts['host']</warning>;
}

function arrayOnlyCallCaught($url)
{
    $parts = parse_url($url);
    try {
        $size = count($parts);
    } catch (TypeError $error) {
        $size = 0;
    }
    return <warning descr="Reject URL parsing failure before indexing.">$parts['host']</warning>;
}

function testedInAnotherScope($url)
{
    $parts = parse_url($url);
    $check = function () use ($parts) {
        if ($parts === false) {
            fail('bad url');
        }
    };
    return <warning descr="Reject URL parsing failure before indexing.">$parts['host']</warning>;
}
