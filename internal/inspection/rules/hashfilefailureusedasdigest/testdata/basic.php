<?php
<warning descr="Reject file hashing failure before using the digest.">file_put_contents($manifest, hash_file('sha256', $path))</warning>;

function guardOnOtherValue(string $path, string $manifest, $other): int
{
    $digest = hash_file('sha256', $path);
    if (!$other) {
        return -1;
    }
    return (int) <warning descr="Reject file hashing failure before using the digest.">file_put_contents($manifest, $digest)</warning>;
}

function guardBeforeReassignment(string $path, string $manifest): int
{
    $digest = hash_file('sha256', $path);
    if (!$digest) {
        return -1;
    }
    $digest = hash_file('md5', $path);
    return (int) <warning descr="Reject file hashing failure before using the digest.">file_put_contents($manifest, $digest)</warning>;
}

function wrongBranch(string $path): bool
{
    $digest = hash_file('sha256', $path);
    return $digest === false || $manifest === '' ? <warning descr="Reject file hashing failure before using the digest.">$digest === 'abc'</warning> : false;
}
