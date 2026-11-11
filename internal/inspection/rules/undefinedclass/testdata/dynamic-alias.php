<?php
// A class_alias() with computed names may declare any global name
// (framework facades); namespaced names are still checked.
foreach (['Route' => 'App\Routing\Router'] as $alias => $class) {
    class_alias($class, $alias);
}
\Route::get('/');
<error descr="Fix the name or install the package declaring App\Routing\Missing.">\App\Routing\Missing</error>::get('/');
