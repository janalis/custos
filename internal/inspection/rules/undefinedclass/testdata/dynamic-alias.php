<?php
// A class_alias() with computed names may declare any name, global
// (framework facades) or namespaced.
foreach (['Route' => 'App\Routing\Router'] as $alias => $class) {
    class_alias($class, $alias);
}
\Route::get('/');
\App\Routing\Missing::get('/');
