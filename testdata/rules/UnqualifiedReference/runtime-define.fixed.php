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
    echo \E_PARSE, \E_DEPRECATED, E_CORE_ERROR;
    echo \E_CORE_WARNING, E_COMPILE_ERROR, \E_USER_NOTICE;
}

namespace {
    \define(__NAMESPACE__ . '\M_PI', 1);
    echo \M_PI;
}

namespace App\Other {
    echo \E_ALL, E_NOTICE, \M_PI;
}
