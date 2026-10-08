<?php
// Failure markers of calls, @-silenced calls, invokable objects, int as float.
/** @param string $data */
function conv($data, string $charset)
{
    $data = @iconv($charset, 'UTF-8//IGNORE', $data);
    if ($data === false) {
        return '';
    }
    return $data;
}

/** @param int|null $ts */
function stamp(string $p, $ts = null)
{
    if ($ts === null) {
        $ts = filemtime($p);
    }
    return $ts;
}

/** @param array $files */
function draft($files)
{
    $files = lookupFiles();
    return $files;
}

/** @return array|bool|null */
function lookupFiles()
{
    return null;
}

/** @param bool|null $flag */
function flags($flag)
{
    $flag = maybeFlag();
    return $flag;
}

/** @return bool|null */
function maybeFlag()
{
    return null;
}

class Handler
{
    public function __invoke($r)
    {
        return $r;
    }
}

class Plain
{
}

class Factory
{
    public static function build(): Handler
    {
        return new Handler();
    }

    public static function plain(): Plain
    {
        return new Plain();
    }

    /** @return \Missing\Thing */
    public static function unknown()
    {
        return null;
    }
}

function run(?callable $handler = null, ?callable $cb = null, ?callable $other = null)
{
    if (!$handler) {
        $handler = Factory::build();
    }
    if (!$cb) {
        $cb = Factory::unknown();
    }
    if (!$other) {
        $other = <warning descr="Assigning a value of type \Plain does not match the parameter's declared type.">Factory::plain()</warning>;
    }
    return [$handler(1), $cb(2), $other];
}

function nearest(float $grade, string $s)
{
    if ($grade < 1) {
        $grade = 1;
    }
    $s = <warning descr="Assigning a value of type int does not match the parameter's declared type.">1</warning>;
    return [$grade, $s];
}
