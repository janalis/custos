<?php
function make_iv(int $n): string
{
    return random_bytes($n);
}
function weak_iv(): string
{
    return str_repeat('0', 16);
}
class IvSource
{
    public function iv(): string { return openssl_random_pseudo_bytes(16); }
    public static function make(): string { return random_bytes(16); }
    public function weak(): string { return 'x'; }
}
