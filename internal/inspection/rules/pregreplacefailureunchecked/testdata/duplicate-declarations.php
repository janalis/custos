<?php
// Copies of a declaration count when every copy takes the text the same way.
if ($legacy) {
    function takeText(string $s) {}
    class Parser { function parse(string $s) {} }
    function takeAny(string $s) {}
} else {
    function takeText(string $s) {}
    class Parser { function parse(string $s) {} }
    function takeAny($s) {}
}

function sameFunction($path) { <warning descr="Handle regex replacement failure before using its result.">takeText(preg_replace('/a/', 'b', $path))</warning>; }
function sameMethod($path, Parser $parser) { <warning descr="Handle regex replacement failure before using its result.">$parser->parse(preg_replace('/a/', 'b', $path))</warning>; }
function differentCopies($path) { takeAny(preg_replace('/a/', 'b', $path)); }
