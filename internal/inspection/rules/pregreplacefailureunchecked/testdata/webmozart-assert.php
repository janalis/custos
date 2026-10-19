<?php
// webmozart/assert assertions that pass when the replacement failed, matched by the class name written.
namespace App {
    use Webmozart\Assert\Assert;
    use Webmozart\Assert\Assert as Check;

    function looselyEmpty($p) { <warning descr="Handle regex replacement failure before using its result.">Assert::eq(preg_replace('/a/', 'b', $p), '')</warning>; }
    function looselyText($p) { Assert::eq(preg_replace('/a/', 'b', $p), '0'); }
    function empty($p) { <warning descr="Handle regex replacement failure before using its result.">Assert::isEmpty(preg_replace('/a/', 'b', $p))</warning>; }
    function aliased($p) { <warning descr="Handle regex replacement failure before using its result.">Check::isEmpty(preg_replace('/a/', 'b', $p))</warning>; }
    function nullableEq($p) { <warning descr="Handle regex replacement failure before using its result.">Assert::nullOrEq(preg_replace('/a/', 'b', $p), null)</warning>; }
    function nullableString($p) { <warning descr="Handle regex replacement failure before using its result.">Assert::nullOrString(preg_replace('/a/', 'b', $p))</warning>; }
    function typed($p) { Assert::string(preg_replace('/a/', 'b', $p)); }
    function dynamic($p, $class) { $class::isEmpty(preg_replace('/a/', 'b', $p)); }
    function spread($p, array $o) { Assert::eq(...$o); }
}
namespace Own {
    class Assert { static function isEmpty($v) {} }
    function project($p) { Assert::isEmpty(preg_replace('/a/', 'b', $p)); }
}
