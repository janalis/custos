<?php
// webmozart/assert assertions that pass when the read failed, matched by the class name written.
namespace App {
    use Webmozart\Assert\Assert;
    use Webmozart\Assert\Assert as Check;

    function looselyEmpty($p) { <warning descr="Handle file reading failure before using its result.">Assert::eq(file_get_contents($p), '')</warning>; }
    function looselyText($p) { Assert::eq(file_get_contents($p), 'x'); }
    function empty($p) { <warning descr="Handle file reading failure before using its result.">Assert::isEmpty(file_get_contents($p))</warning>; }
    function aliased($p) { <warning descr="Handle file reading failure before using its result.">Check::isEmpty(file_get_contents($p))</warning>; }
    function nullableEq($p) { <warning descr="Handle file reading failure before using its result.">Assert::nullOrEq(file_get_contents($p), null)</warning>; }
    function nullableString($p) { Assert::nullOrString(file_get_contents($p)); }
    function typed($p) { Assert::string(file_get_contents($p)); }
    function dynamic($p, $class) { $class::isEmpty(file_get_contents($p)); }
    function spread($p, array $o) { Assert::eq(...$o); }
}
namespace Own {
    class Assert { static function isEmpty($v) {} }
    function project($p) { Assert::isEmpty(file_get_contents($p)); }
}
