<?php
function collect($handler = 'func_get_args')
{
    $grab = <warning descr="'func_get_args' reads the caller scope and cannot be invoked indirectly since PHP 7.1.">$handler</warning>();

    $rows = array_map(<warning descr="'extract' reads the caller scope and cannot be invoked indirectly since PHP 7.1.">'extract'</warning>, $input);
    call_user_func_array(<warning descr="'compact' reads the caller scope and cannot be invoked indirectly since PHP 7.1.">'\compact'</warning>, ['a', 'b']);
    array_walk($list, <warning descr="'parse_str' reads the caller scope and cannot be invoked indirectly since PHP 7.1.">"parse_str"</warning>);

    $reader = 'mb_parse_str';
    <warning descr="'mb_parse_str' reads the caller scope and cannot be invoked indirectly since PHP 7.1.">$reader</warning>($query, $out);
}
