<?php
// A file the function made with tempnam() is cleaned up, not deleted for
// the caller.
function inodesLeft(string $dir) {
    $probe = @tempnam($dir, 'probe');
    if (false === $probe) {
        return false;
    }
    unlink($probe);
    return true;
}
function namedScratch() {
    $tmp = tempnam(sys_get_temp_dir(), 'x');
    unlink(filename: $tmp);
    return true;
}
// Once the variable names another file, its deletion is the caller's.
function replaced(string $other) {
    $tmp = tempnam(sys_get_temp_dir(), 'x');
    $tmp = $other;
    <warning descr="Return the deletion result or handle failure.">unlink($tmp)</warning>;
    return true;
}
function notMade(string $given) {
    <warning descr="Return the deletion result or handle failure.">unlink($given)</warning>;
    return true;
}
function spread(array $paths) {
    <warning descr="Return the deletion result or handle failure.">unlink(...$paths)</warning>;
    return true;
}
