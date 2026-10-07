<?php
function route($req, $list) {
    if ($req === null) { exit(3); };
    if (!$list) { throw new LogicException('none'); };
    foreach ($list as $item) {
        if ($item < 0) { continue; }
        if ($item > 9) { break; }
    }
    if ($req === 'a') { return 1; }
    if ($req === 'b') { $list = []; } else { $list = [1]; }
    if ($req === 'x') { die; }
    $list[] = 1; $list[] = 2;
    if ($req === 'y') { return 0; }
    if($req === 'z'){ $list = 1; }  elseif($req) { $list = 2; }
    if ($req === 'w') { return; }
    return $list;
}
