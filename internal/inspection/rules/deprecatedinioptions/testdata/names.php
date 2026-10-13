<?php
namespace Config {
    function ini_set($k, $v) {}
    ini_set('safe_mode', '0');
    \INI_SET(<warning descr="Ini directive 'safe_mode' no longer exists since PHP 5.4.0.">'safe_mode'</warning>, '0');
    Ini_Get(<warning descr="Ini directive 'register_globals' no longer exists since PHP 5.4.0.">'register_globals'</warning>);
}
