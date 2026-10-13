<?php

function unchangedLocal($data)
{
    $level = 20;
    helper();
    return <error descr="Use a supported compression level.">gzencode($data, $level)</error>;
}

function copiedLocal($data)
{
    $first = 20;
    $second = $first;
    $first = $second;
    return <error descr="Use a supported compression level.">gzcompress($data, ($second))</error>;
}

function arrowWritesItsOwnCopy($data)
{
    $level = 20;
    $lower = fn() => $level--;
    return <error descr="Use a supported compression level.">gzdeflate($data, $level)</error>;
}

function importedByValue($data)
{
    $level = 20;
    $reader = function () use ($level) {
        return $level;
    };
    return <error descr="Use a supported compression level.">gzencode($data, $level)</error>;
}

function unrelatedCounter($data)
{
    $level = 20;
    $rounds = 0;
    $rounds++;
    function declaredInside()
    {
        $level = 1;
        $level--;
        extract(array());
    }
    return <error descr="Use a supported compression level.">gzencode($data, $level)</error>;
}

function sharedWithoutCalls($data, &$out)
{
    global $configured;
    $configured = 20;
    $out = 15;
    return <error descr="Use a supported compression level.">gzencode($data, $configured)</error> . <error descr="Use a supported compression level.">gzencode($data, $out)</error>;
}

class Archiver
{
    public function pack($data)
    {
        $level = 20;
        $this->prepare();
        return <error descr="Use a supported compression level.">gzencode($data, $level)</error>;
    }
}

$packed = 'sample';
$callback = function () use (&$shared, $packed) {
    $shared = 20;
    $local = 20;
    $size = strlen($packed);
    return <error descr="Use a supported compression level.">gzencode($packed, $shared)</error> . <error descr="Use a supported compression level.">gzencode($packed, $local)</error>;
};
