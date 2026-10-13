<?php
$enc = ini_get(<warning descr="Ini directive 'mbstring.http_output' is deprecated since PHP 5.6.0; use default_charset.">'mbstring.http_output'</warning>);
ini_set(<warning descr="Ini directive 'safe_mode' no longer exists since PHP 5.4.0.">"SAFE_MODE"</warning>, '0');
ini_alter(<warning descr="Ini directive 'mbstring.script_encoding' no longer exists since PHP 5.4.0; use zend.script_encoding.">'mbstring.script_encoding'</warning>, 'UTF-8');
\ini_restore(<warning descr="Ini directive 'always_populate_raw_post_data' is deprecated since PHP 5.6.0.">'always_populate_raw_post_data'</warning>);
ini_set('asp_tags', '1');
ini_set('track_errors', '1');
ini_set('memory_limit', '64M');
$key = 'safe_mode';
ini_get($key);
ini_get_all('safe_mode');
$cfg->ini_set('safe_mode', 1);
