<?php
function mixedCase($items, $row)
{
    array_map(<warning descr="'Compact' reads the caller scope and cannot be invoked indirectly since PHP 7.1.">'Compact'</warning>, $items);
    Array_Filter($items, <warning descr="'EXTRACT' reads the caller scope and cannot be invoked indirectly since PHP 7.1.">'EXTRACT'</warning>);
    CALL_USER_FUNC(<warning descr="'Func_Get_Args' reads the caller scope and cannot be invoked indirectly since PHP 7.1.">'\Func_Get_Args'</warning>);
    $grab = 'Get_Defined_Vars';
    <warning descr="'Get_Defined_Vars' reads the caller scope and cannot be invoked indirectly since PHP 7.1.">$grab</warning>();
    array_map('Trim', $items);
}
