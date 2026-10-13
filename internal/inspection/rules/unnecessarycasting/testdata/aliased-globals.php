<?php
$total = 0;
$named = 0;
$dynamic = '';
$local = 0;
$kept = 0.5;

function loadTotal() { global $total; $total = '6 limes'; }
function loadNamed() { $GLOBALS['named'] = '3 kiwis'; }
class Loader { public function run($key) { return $GLOBALS['other']; } }

loadTotal();
loadNamed();
fillLocal($local);

return [(int) $total, (int) $named, (int) $local, <weak_warning descr="Operand already has the target type; remove the cast.">(string)</weak_warning> $dynamic, <weak_warning descr="Operand already has the target type; remove the cast.">(float)</weak_warning> $kept];
