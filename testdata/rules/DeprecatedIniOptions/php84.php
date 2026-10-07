<?php
ini_set(<warning descr="Ini directive 'track_errors' no longer exists since PHP 8.0.0.">'track_errors'</warning>, '1');
ini_get(<warning descr="Ini directive 'mbstring.func_overload' no longer exists since PHP 8.0.0.">'mbstring.func_overload'</warning>);
ini_get(<warning descr="Ini directive 'pdo_odbc.db2_instance_name' no longer exists since PHP 8.0.0.">'pdo_odbc.db2_instance_name'</warning>);
ini_set(<warning descr="Ini directive 'assert.quiet_eval' no longer exists since PHP 8.0.0.">'assert.quiet_eval'</warning>, '1');
ini_set(<warning descr="Ini directive 'log_errors_max_len' no longer exists since PHP 8.1.0.">'log_errors_max_len'</warning>, '0');
ini_get(<warning descr="Ini directive 'mysqlnd.fetch_data_copy' no longer exists since PHP 8.1.0.">'mysqlnd.fetch_data_copy'</warning>);
ini_set(<warning descr="Ini directive 'auto_detect_line_endings' is deprecated since PHP 8.1.0.">'auto_detect_line_endings'</warning>, '1');
ini_get(<warning descr="Ini directive 'date.default_latitude' is deprecated since PHP 8.1.0.">'date.default_latitude'</warning>);
ini_get(<warning descr="Ini directive 'date.sunset_zenith' is deprecated since PHP 8.1.0.">'Date.Sunset_Zenith'</warning>);
ini_get(<warning descr="Ini directive 'filter.default' is deprecated since PHP 8.1.0.">'filter.default'</warning>);
ini_set(<warning descr="Ini directive 'assert.active' is deprecated since PHP 8.3.0; use zend.assertions.">'assert.active'</warning>, '1');
ini_set(<warning descr="Ini directive 'assert.exception' is deprecated since PHP 8.3.0.">'assert.exception'</warning>, '1');
ini_get(<warning descr="Ini directive 'opcache.consistency_checks' no longer exists since PHP 8.3.0.">'opcache.consistency_checks'</warning>);
ini_set(<warning descr="Ini directive 'session.sid_length' is deprecated since PHP 8.4.0.">'session.sid_length'</warning>, '32');
ini_get(<warning descr="Ini directive 'session.sid_bits_per_character' is deprecated since PHP 8.4.0.">'session.sid_bits_per_character'</warning>);
ini_get(<warning descr="Ini directive 'allow_url_include' is deprecated since PHP 7.4.0.">'allow_url_include'</warning>);
ini_get('zend.assertions');
ini_set('session.use_strict_mode', '1');
// not reported: no argument / first-class callable
ini_get();
$get = ini_get(...);
