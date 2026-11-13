<?php
// Below PHP 8.0 builtins are not checked; the project's functions are.
strlen();
strlen('a', 'b');
function total($items, $tax) {}
<error descr="Pass the required argument $tax.">total([])</error>;

// A polyfill of a function PHP 8.0 declares: the builtin runs on newer
// versions, so the polyfill's parameters are not the contract.
function str_contains($haystack, $needle) {}
str_contains('a');
