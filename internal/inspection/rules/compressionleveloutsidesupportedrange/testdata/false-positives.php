<?php

function postDecrement($data)
{
    $level = 10;
    $level--;
    return gzencode($data, $level);
}

function preDecrement($data)
{
    $level = 10;
    --$level;
    return gzencode($data, $level);
}

function copiedAfterDecrement($data)
{
    $first = 10;
    $first--;
    $second = $first;
    return gzencode($data, $second);
}

function clampedByReferenceClosure($data)
{
    $level = 20;
    $clamp = function () use (&$level) {
        $level = min(9, $level);
    };
    $clamp();
    return gzencode($data, $level);
}

function extracted($data)
{
    $level = 20;
    extract(array('level' => 1));
    return gzencode($data, $level);
}

function dynamicDecrement($data)
{
    $level = 10;
    $name = 'level';
    $$name--;
    return gzencode($data, $level);
}

function dynamicCompound($data)
{
    $level = 10;
    ${'level'} -= 5;
    return gzencode($data, $level);
}

function globalRewrittenByCall($data)
{
    global $configured;
    $configured = 20;
    resetConfiguredLevel();
    return gzencode($data, $configured);
}

function resetConfiguredLevel()
{
    $GLOBALS['configured'] = 5;
}

function staticRewrittenByRecursion($data, $depth)
{
    static $level;
    $level = 20;
    if ($depth > 0) {
        staticRewrittenByRecursion($data, $depth - 1);
    }
    $level = $depth > 0 ? 20 : 3;
    return $depth;
}

function staticAcrossCall($data)
{
    static $level;
    $level = 20;
    lowerStaticLevel();
    return gzencode($data, $level);
}

function referenceParameterAcrossCall($data, &$level)
{
    $level = 20;
    $data = new Normaliser($data);
    return gzencode($data->text(), $level);
}

$referenced = function () use (&$shared) {
    $shared = 20;
    Registry::reset();
    return gzencode('sample', $shared);
};
