<?php
function safe_first(string $a, string $b, $choose) {
    $v = gzdecode(gzencode($a));
    if ($choose) { $v = gzdecode($b); }
    <warning descr="Check decompression before consuming data.">strlen($v)</warning>;
}
function unsafe_first(string $a, string $b, $choose) {
    $v = gzdecode($a);
    if ($choose) { $v = gzdecode(gzencode($b)); }
    <warning descr="Check decompression before consuming data.">strlen($v)</warning>;
}
function both_safe(string $a, string $b, $choose) {
    $v = gzdecode(gzencode($a));
    if ($choose) { $v = gzdecode(gzencode($b)); }
    strlen($v);
}
