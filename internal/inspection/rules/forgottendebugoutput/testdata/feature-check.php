<?php
function trace(int $offset): array
{
    if (\function_exists('xdebug_get_function_stack') && $trace = @xdebug_get_function_stack()) {
        return $trace;
    }
    if (function_exists('\\XDEBUG_GET_FUNCTION_STACK')) {
        return xdebug_get_function_stack();
    } elseif (function_exists('xdebug_get_declared_vars')) {
        xdebug_get_declared_vars();
    }
    if (\function_exists('xdebug_get_function_stack') and $ok) {
        xdebug_get_function_stack();
    }
    $stack = function_exists('xdebug_get_function_stack') ? xdebug_get_function_stack() : [];
    $other = function_exists('xdebug_info') ? <error descr="Debug output call; remove it if it was left over from debugging.">xdebug_get_function_stack()</error> : [];
    if (function_exists('xdebug_get_function_stack') || $force) {
        <error descr="Debug output call; remove it if it was left over from debugging.">xdebug_get_function_stack()</error>;
    }
    if (function_exists($name)) {
        <error descr="Debug output call; remove it if it was left over from debugging.">xdebug_get_declared_vars()</error>;
    }
    $x = $ready ?: <error descr="Debug output call; remove it if it was left over from debugging.">xdebug_get_declared_vars()</error>;
    return $stack;
}
