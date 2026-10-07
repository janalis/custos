<?php
// compact() and get_defined_vars() read variables by name.
function render(string $template, $find) {
    if (!($page = $find())) {
        throw new \RuntimeException();
    }
    $nav = [];
    return [$template, compact('nav', ['page'])];
}

function dump(array $context) {
    $context['seen'] = true;
    return get_defined_vars();
}

function lost(array $context) {
    <weak_warning descr="Value is only written here and never read; the write is lost.">$context['seen']</weak_warning> = true;
    return compact('other');
}

$cb = function () use ($config, &$state) {
    return compact('config', 'state');
};
