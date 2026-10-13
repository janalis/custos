<?php
function route($req, $list) {
    if ($req === 'c') { $list[] = 2; } else { ; }
    if ($req === 'd') return 4; else ;
    if ($req === 'e') { return 5; } else $list = [];
    if ($req === 'f') { ; } else if ($req === 'g') { return 6; } else { ; }
    if ($req === 'h'):
        return 7;
    else:
        $list = null;
    endif;
    if ($req === 'i') { return 8; } elseif ($req === 'j') $list = 3;
    if ($req === 'k') { if ($list) { return 9; } } else { $list = 4; }
    if ($req === 'l') {} else { $list = 5; }
    if ($req === 'm') { return 1; }
    return $list;
}
