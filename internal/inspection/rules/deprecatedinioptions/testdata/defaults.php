<?php
// Directives deprecated only for a value other than their default.
ini_get('assert.active');
ini_restore('session.sid_length');
ini_set('assert.active', 1);
ini_set('assert.exception', true);
ini_alter('assert.warning', 'On');
ini_set('assert.bail', 0);
ini_set('assert.bail', false);
ini_set('assert.bail', 'no');
ini_set('assert.callback', null);
ini_set('assert.callback', '');
ini_set('session.sid_length', '32');
ini_set('session.sid_bits_per_character', 4);
ini_set('session.sid_length', value: 0x20);
// Other values, or values not known, are reported.
ini_set(<warning descr="Ini directive 'assert.active' is deprecated since PHP 8.3.0; use zend.assertions.">'assert.active'</warning>, '2x');
ini_set(<warning descr="Ini directive 'assert.warning' is deprecated since PHP 8.3.0.">'assert.warning'</warning>, "\x30");
ini_set(<warning descr="Ini directive 'assert.warning' is deprecated since PHP 8.3.0.">'assert.warning'</warning>, 1.0);
ini_set(<warning descr="Ini directive 'assert.warning' is deprecated since PHP 8.3.0.">'assert.warning'</warning>, PHP_INT_MAX);
ini_set(<warning descr="Ini directive 'assert.bail' is deprecated since PHP 8.3.0.">'assert.bail'</warning>, 'yes');
ini_set(<warning descr="Ini directive 'assert.callback' is deprecated since PHP 8.3.0.">'assert.callback'</warning>, 'cb');
ini_set(<warning descr="Ini directive 'assert.callback' is deprecated since PHP 8.3.0.">'assert.callback'</warning>, $cb);
ini_set(<warning descr="Ini directive 'session.sid_length' is deprecated since PHP 8.4.0.">'session.sid_length'</warning>, 'x');
ini_set(<warning descr="Ini directive 'session.sid_length' is deprecated since PHP 8.4.0.">'session.sid_length'</warning>);
ini_set(<warning descr="Ini directive 'session.sid_length' is deprecated since PHP 8.4.0.">'session.sid_length'</warning>, ...$values);
ini_set(<warning descr="Ini directive 'session.sid_length' is deprecated since PHP 8.4.0.">'session.sid_length'</warning>, other: 32);
