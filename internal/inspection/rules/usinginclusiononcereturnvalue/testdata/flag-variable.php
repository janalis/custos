<?php
function loadDriver(array $dirs, string $file)
{
    $found = false;
    foreach ($dirs as $dir) {
        $found = @include_once $dir . $file;
        if ($found) {
            break;
        }
    }
    return $found ? 'ok' : 'missing';
}

function loadOrDie(string $file)
{
    $res = include_once $file;
    if (!$res) {
        die('cannot load');
    }
}

function loadAndForget(string $file)
{
    $res = require_once $file;
}

function loadFlagInCondition(string $file)
{
    if (($ok = include_once $file) && $ok) {
        return true;
    }
    return false;
}

function loadConfig(string $file)
{
    $config = <error descr="Only the first include_once/require_once returns the file's value; later ones return true.">require_once $file</error>;
    if (!$config) {
        return [];
    }
    return $config['secret'];
}

function loadAppended(string $file)
{
    $all = <error descr="Only the first include_once/require_once returns the file's value; later ones return true.">include_once $file</error>;
    $all[] = 'x';
    $all .= 'y';
}

function loadAssignedUse(string $file)
{
    echo $v = <error descr="Only the first include_once/require_once returns the file's value; later ones return true.">include_once $file</error>;
    $w = &$v;
    $$file = <error descr="Only the first include_once/require_once returns the file's value; later ones return true.">include_once $file</error>;
    $x .= <error descr="Only the first include_once/require_once returns the file's value; later ones return true.">include_once $file</error>;
    $y = &<error descr="Only the first include_once/require_once returns the file's value; later ones return true.">include_once $file</error>;
    $o->p = <error descr="Only the first include_once/require_once returns the file's value; later ones return true.">include_once $file</error>;
}

$app = <error descr="Only the first include_once/require_once returns the file's value; later ones return true.">require_once __DIR__ . '/bootstrap/app.php'</error>;
$app->run();
$loaded = include_once 'helpers.php';
if (!$loaded) {
    exit(1);
}

// A variable other code can reach is not a private success flag.
function loadGlobal(string $file)
{
    global $config;
    $config = <error descr="Only the first include_once/require_once returns the file's value; later ones return true.">include_once $file</error>;
}

function loadStatic(string $file)
{
    static $cache;
    $cache = <error descr="Only the first include_once/require_once returns the file's value; later ones return true.">include_once $file</error>;
}

function loadInto(string $file, &$out)
{
    $out = <error descr="Only the first include_once/require_once returns the file's value; later ones return true.">include_once $file</error>;
}

function loadIntoCopy(string $file, &$out, $copy)
{
    $copy = include_once $file;
}

function loadCompact(string $file)
{
    $x = <error descr="Only the first include_once/require_once returns the file's value; later ones return true.">include_once $file</error>;
    return compact('x');
}

function loadDefinedVars(string $file)
{
    $x = <error descr="Only the first include_once/require_once returns the file's value; later ones return true.">include_once $file</error>;
    return \get_defined_vars();
}

function loadVariableVariable(string $file, string $name)
{
    $x = <error descr="Only the first include_once/require_once returns the file's value; later ones return true.">include_once $file</error>;
    if ($x) {
        return $$name;
    }
}

function loadEval(string $file, string $code)
{
    $x = <error descr="Only the first include_once/require_once returns the file's value; later ones return true.">include_once $file</error>;
    eval($code);
}

function loadForTemplate(string $file)
{
    $data = <error descr="Only the first include_once/require_once returns the file's value; later ones return true.">include_once $file</error>;
    include 'template.php';
}

function loadWithFallback(string $a, string $b)
{
    $res = @include_once $a;
    if (!$res) {
        $res = @include_once $b;
    }
    $helper = function ($name) {
        return compact('name') + [$$name];
    };
    return $helper;
}

$byRef = function (string $file) use ($plain, &$shared) {
    $shared = <error descr="Only the first include_once/require_once returns the file's value; later ones return true.">include_once $file</error>;
};
$byValue = function (string $file) use (&$shared, $plain) {
    $plain = include_once $file;
};
