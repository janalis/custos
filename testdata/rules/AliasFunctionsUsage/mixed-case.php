<?php
namespace Billing {
    function Chop($s) { return $s; }

    $a = CHOP($label);
    $b = \<warning descr="Use 'rtrim(...)' instead of the alias 'CHOP(...)'.">CHOP</warning>($label);
    $total = <warning descr="Use 'count(...)' instead of the alias 'SizeOf(...)'.">SizeOf</warning>($rows);
    $csv = <warning descr="Use 'implode(...)' instead of the alias 'JOIN(...)'.">JOIN</warning>(',', $cells);
    <warning descr="'Magic_Quotes_Runtime(...)' is a legacy alias (deprecated 5.3, removed 7.0); stop relying on it.">Magic_Quotes_Runtime</warning>(0);
    $m = $repo->SizeOf();
}
