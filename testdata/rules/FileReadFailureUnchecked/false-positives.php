<?php

function stop(): never
{
    throw new RuntimeException('unreadable');
}

function alternativeExits($path)
{
    $text = file_get_contents($path) or die('unreadable');

    return strlen($text);
}

function alternativeThrows($path)
{
    ($text = file_get_contents($path)) || stop();

    return strlen($text);
}

function looseGuard($path)
{
    $text = file_get_contents($path);
    if ($text == false) {
        return 0;
    }

    return strlen($text);
}

function looseBranches($path)
{
    $text = file_get_contents($path);
    if ($text != false) {
        return strlen($text);
    } elseif ($path === '') {
        return -1;
    }

    return $text == false ? 0 : strlen($text);
}

function looseElse($path)
{
    $text = file_get_contents($path);
    if (null == $text) {
        $size = 0;
    } elseif ($path === '') {
        $size = strlen($text);
    } else {
        $size = strlen($text) + 1;
    }

    return $size;
}

function shortCircuit($handle)
{
    $text = stream_get_contents($handle);

    return $text != false && strlen($text) > 3;
}

function neverHelper($path)
{
    $text = file_get_contents($path);
    if ($text === false) {
        error_log('unreadable');
        stop();
    }

    return strlen($text);
}

function neverHelperVariants($path)
{
    $text = file_get_contents($path);
    if (!is_string($text) || $path === '')
        stop();
    $copy = file_get_contents($path . '.bak');
    if (empty($copy)) {
        stop();
    }

    return strlen($text) + strlen($copy);
}

function assignedInCondition($path)
{
    if (($text = file_get_contents($path)) == false) {
        return 0;
    }
    if ($path !== '')
        return strlen($text);

    return 1;
}
