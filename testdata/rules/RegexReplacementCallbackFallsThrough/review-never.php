<?php
function reject(string $why): never { throw new DomainException($why); }
function remember(string $what): void { error_log($what); }

echo preg_replace_callback('~[a-z]~', function ($m) {
    if ($m[0] !== 'q') {
        return strtoupper($m[0]);
    }
    reject('letter q is not supported');
}, 'abc');

echo preg_replace_callback('~[a-z]~', function ($m) {
    if ($m[0] !== 'q') {
        return strtoupper($m[0]);
    } else {
        reject('letter q is not supported');
    }
}, 'abc');

echo preg_replace_callback('~[a-z]~', function ($m) {
    helper_declared_elsewhere($m[0]);
}, 'abc');

echo <warning descr="Return the replacement from the callback.">preg_replace_callback('~[a-z]~', function ($m) {
    if ($m[0] !== 'q') {
        return strtoupper($m[0]);
    }
    remember('letter q was skipped');
}, 'abc')</warning>;
