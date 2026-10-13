<?php
namespace Billing {
    function Chop($s) { return $s; }

    $a = CHOP($label);
    $b = \rtrim($label);
    $total = count($rows);
    $csv = implode(',', $cells);
    Magic_Quotes_Runtime(0);
    $m = $repo->SizeOf();
}
