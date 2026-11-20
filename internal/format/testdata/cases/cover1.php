<?php
echo 0; // presets: per symfony; set: open_tag_own_line=false, import_order=length, remove_unused_imports=true, yoda=yoda, array_syntax=short
use Aa\Bb;
use Bb\Aa;
use Unused\X /* c */;
function f() {
    // @custos-format-ignore-next
    $a   =   1;
    $f = function () { $a = f(1); $b = 2; };
    $g = function () { return 1;
    };
    return ($a);
}
$x = array /* c */ (1);
if ($y == array()) {}
// @custos-format-ignore-next
if ($z == 1) {}
echo (new Aa) . new Bb;
