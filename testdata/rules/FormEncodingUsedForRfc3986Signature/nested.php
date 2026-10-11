<?php
function signRfc3986($key, $extra)
{
    $byKey = <warning descr="Use RFC3986 query encoding for this signature protocol.">http_build_query(["full name" => "ab"])</warning>;
    $a = hash_hmac("sha256", $byKey, $key);
    $byNested = <warning descr="Use RFC3986 query encoding for this signature protocol.">http_build_query(["tags" => ["red", "dark blue"]])</warning>;
    $b = hash_hmac("sha256", $byNested, $key);
    $byNestedKey = <warning descr="Use RFC3986 query encoding for this signature protocol.">http_build_query(["filter" => ["sort by" => 1]])</warning>;
    $c = hash_hmac("sha256", $byNestedKey, $key);
    // A null or empty value drops the whole pair, spaced key included.
    $dropped = http_build_query(["full name" => null, "more tags" => [], "other field" => $extra]);
    $d = hash_hmac("sha256", $dropped, $key);
    return $a . $b . $c . $d;
}
