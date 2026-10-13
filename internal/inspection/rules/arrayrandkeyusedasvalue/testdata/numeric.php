<?php
// Key 1 matches no case, so the consumer still misses a value.
function level($c) { switch ($c) { case '0': return 'zero'; case '2': return 'two'; } }
echo level(<warning descr="Read the array element at the returned random key.">array_rand(['0', '2'])</warning>);
function mixed($c) { switch ($c) { case '0': return 'zero'; case 'one': return 'one'; } }
echo mixed(<warning descr="Read the array element at the returned random key.">array_rand(['0', 'one'])</warning>);
