<?php
namespace App\Poly {
    \define(__NAMESPACE__ . '\E_ALL', 7);
    \define('App\\Poly\\E_ERROR', 8);
    \define(__NAMESPACE__ . '\\' . 'E_WARNING', 9);
    \define($prefix . '\E_NOTICE', 10);
    \define(('\app\Poly\\') . "E_STRICT", 11);
    \define(constant_name: __namespace__ . '\E_USER_ERROR', value: 12);
    \define('Other\\E_PARSE', 1);
    \define('PRE_' . 'E_DEPRECATED', 1);
    \define($prefix . 'E_CORE_ERROR', 1);
    \define('\E_CORE_WARNING' . $suffix, 1);
    \define(__LINE__ . '\E_COMPILE_ERROR', 1);
    \define($a + '\E_USER_NOTICE', 1);
    \define(...$pair);
    $definer('x', 1);

    echo E_ALL, E_ERROR, E_WARNING, E_NOTICE, E_STRICT, E_USER_ERROR;
    echo <weak_warning descr="Write '\E_PARSE' to allow compile-time binding.">E_PARSE</weak_warning>, <weak_warning descr="Write '\E_DEPRECATED' to allow compile-time binding.">E_DEPRECATED</weak_warning>, E_CORE_ERROR;
    echo <weak_warning descr="Write '\E_CORE_WARNING' to allow compile-time binding.">E_CORE_WARNING</weak_warning>, E_COMPILE_ERROR, <weak_warning descr="Write '\E_USER_NOTICE' to allow compile-time binding.">E_USER_NOTICE</weak_warning>;
}

namespace {
    \define(__NAMESPACE__ . '\M_PI', 1);
    echo <weak_warning descr="Write '\M_PI' to allow compile-time binding.">M_PI</weak_warning>;
}

namespace App\Other {
    echo <weak_warning descr="Write '\E_ALL' to allow compile-time binding.">E_ALL</weak_warning>, E_NOTICE, <weak_warning descr="Write '\M_PI' to allow compile-time binding.">M_PI</weak_warning>;
}
