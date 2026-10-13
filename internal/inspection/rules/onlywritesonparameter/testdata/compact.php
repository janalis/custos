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

// Names held in a variable (phpBB's event idiom) or spread: any variable may
// be read.
function search($d) {
    $sort = $join = '';
    $vars = array('sort', 'join');
    extract($d->trigger('core.search', compact($vars)));
    return $sort;
}

function spread(array $context, array $names) {
    $context['seen'] = true;
    return compact(...$names);
}

function mixedNames(array $context, $more) {
    $context['seen'] = true;
    return compact(['other', $more]);
}

// Variable variables may write or read any variable.
function langs($user) {
    $rule_lang = $action_lang = array();
    array_map(function ($m) use (&$rule_lang, &$action_lang, $user) {
        ${strtolower($m[0]) . '_lang'}[] = $user->lang[$m[1]];
    }, ['x']);
    return [$rule_lang, $action_lang];
}
