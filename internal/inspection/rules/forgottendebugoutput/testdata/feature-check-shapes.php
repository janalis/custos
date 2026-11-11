<?php
function earlyExit(): array
{
    if (!function_exists('xdebug_get_function_stack')) {
        return [];
    }
    return xdebug_get_function_stack();
}

function elseArm(): array
{
    if (!function_exists('xdebug_get_function_stack')) {
        $stack = [];
    } else {
        $stack = xdebug_get_function_stack();
    }
    return $stack;
}

function elseifArm(bool $quiet): void
{
    if ($quiet) {
        return;
    } elseif (!function_exists('xdebug_print_function_stack')) {
        echo 'unavailable';
    } else {
        xdebug_print_function_stack();
    }
}

function shortCircuit(): void
{
    function_exists('xdebug_print_function_stack') || throw new RuntimeException('xdebug missing');
    xdebug_print_function_stack();
    !function_exists('xdebug_get_declared_vars') || xdebug_get_declared_vars();
}

function stored(): void
{
    $probe = function_exists('xdebug_print_function_stack');
    if ($probe) {
        xdebug_print_function_stack();
    }
}

function fallsThrough(): void
{
    if (!function_exists('xdebug_print_function_stack')) {
        echo 'unavailable';
    }
    <error descr="Debug output call; remove it if it was left over from debugging.">xdebug_print_function_stack()</error>;
    if (function_exists('xdebug_print_function_stack')) {
        echo 'available';
    } else {
        <error descr="Debug output call; remove it if it was left over from debugging.">xdebug_print_function_stack()</error>;
    }
}
