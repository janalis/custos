<?php
namespace App;

// A test double of tempnam() declared in the namespace: the unqualified call
// still names PHP's function, so the file is the function's own scratch.
function tempnam($dir, $prefix)
{
    return '/tmp/mocked';
}

function probe()
{
    $tmp = tempnam(sys_get_temp_dir(), 'x');
    unlink($tmp);
    return true;
}

// Naming the namespaced function explicitly is not PHP's tempnam().
function explicit()
{
    $tmp = \App\tempnam(sys_get_temp_dir(), 'x');
    <warning descr="Return the deletion result or handle failure.">unlink($tmp)</warning>;
    return true;
}
