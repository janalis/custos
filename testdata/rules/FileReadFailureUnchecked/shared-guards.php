<?php
function fail(string $message): never
{
    throw new RuntimeException($message);
}

// Failure is ruled out before the use, in forms the shared flow analysis reads.
function leavesOnFailure($file)
{
    $text = file_get_contents($file);
    $text !== false or exit(1);
    return strlen($text);
}

function storedTest($file)
{
    $text = file_get_contents($file);
    $ok = $text !== false;
    if (!$ok) {
        return 0;
    }
    return strlen($text);
}

function neverReturns($file)
{
    $text = file_get_contents($file);
    if ($text === false) {
        fail('unreadable');
    }
    return strlen($text);
}

function continuesAfterElse($file)
{
    $text = file_get_contents($file);
    if ($text == false) {
        return 0;
    } else {
        error_log('read');
    }
    return strlen($text);
}

function retried($files)
{
    for ($text = file_get_contents($files[0]); $text !== false; $text = file_get_contents($files[1])) {
        echo strlen($text);
    }
}

// The evidence does not hold.
function leavesOnSuccess($file)
{
    $text = file_get_contents($file);
    $text === false or exit(1);
    return <warning descr="Handle file reading failure before using its result.">strlen($text)</warning>;
}

function storedTestOverwritten($file, $flag)
{
    $text = file_get_contents($file);
    $ok = $text !== false;
    $ok = $flag;
    if (!$ok) {
        return 0;
    }
    return <warning descr="Handle file reading failure before using its result.">strlen($text)</warning>;
}

function otherResultTested($file, $other)
{
    $text = file_get_contents($file);
    $second = file_get_contents($other);
    $second !== false or exit(1);
    return <warning descr="Handle file reading failure before using its result.">strlen($text)</warning>;
}
