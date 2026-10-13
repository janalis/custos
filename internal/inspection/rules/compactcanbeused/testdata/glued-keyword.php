<?php
function glued($a, $b) { return<weak_warning descr="Replace with 'compact('a', 'b')'.">[</weak_warning>'a'=>$a,'b'=>$b]; }
function shown($a, $b) { echo<weak_warning descr="Replace with 'compact('a', 'b')'.">[</weak_warning>'a' => $a, 'b' => $b] === [] ? 0 : 1; }
