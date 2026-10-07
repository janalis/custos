<?php
function route($req, $list) {
    if ($req === null) { exit(3); } <warning descr="Drop the 'else' and move its code after the 'if'.">else</warning> { ; }
    if (!$list) { throw new LogicException('none'); }
    <warning descr="Drop the 'else' and move its code after the 'if'.">else</warning> { ; }
    foreach ($list as $item) {
        if ($item < 0) { continue; } <warning descr="Drop the 'else' and move its code after the 'if'.">else</warning> if ($item > 9) { break; }
    }
    if ($req === 'a') { return 1; }
    <warning descr="Turn this 'elseif' into a separate 'if'.">elseif</warning> ($req === 'b') { $list = []; } else { $list = [1]; }
    if ($req === 'x') { die; } <warning descr="Drop the 'else' and move its code after the 'if'.">else</warning> { $list[] = 1; $list[] = 2; }
    if($req === 'y'){ return 0; } <warning descr="Turn this 'elseif' into a separate 'if'.">elseif</warning> ($req === 'z')  { $list = 1; }  elseif($req) { $list = 2; }
    if ($req === 'w') { return; } <warning descr="Drop the 'else' and move its code after the 'if'.">else</warning> {}
    return $list;
}
