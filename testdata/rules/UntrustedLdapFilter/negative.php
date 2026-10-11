<?php
ldap_search($ldap,$base,"(uid=".ldap_escape($_GET["user"],"",LDAP_ESCAPE_FILTER).")");
// Filter-context escaping, unproven values and server-generated $_SERVER entries.
function safe1($ldap, $base, $flags, $name, $user) { return ldap_search($ldap, $base, "(uid=" . ldap_escape($_GET["user"]) . ")"); }
function safe2($ldap, $base, $flags, $name, $user) { return ldap_search($ldap, $base, "(uid=" . ldap_escape($_GET["user"], "", 0) . ")"); }
function safe3($ldap, $base, $flags, $name, $user) { return ldap_search($ldap, $base, "(uid=" . ldap_escape($_GET["user"], "", LDAP_ESCAPE_DN | LDAP_ESCAPE_FILTER) . ")"); }
function safe4($ldap, $base, $flags, $name, $user) { return ldap_search($ldap, $base, "(uid=" . ldap_escape($_GET["user"], "", $flags) . ")"); }
function safe5($ldap, $base, $flags, $name, $user) { return ldap_search($ldap, $base, "(uid=" . ldap_escape("admin", "", LDAP_ESCAPE_DN) . ")"); }
function safe6($ldap, $base, $flags, $name, $user) { return ldap_search($ldap, $base, "(uid=" . normalise($_GET["user"]) . ")"); }
function safe7($ldap, $base, $flags, $name, $user) { return ldap_search($ldap, $base, "(uid={$_GET[$name]})"); }
function safe8($ldap, $base, $flags, $name, $user) { return ldap_search($ldap, $base, "(uid={$user})"); }
function safe9($ldap, $base, $flags, $name, $user) { return ldap_search($ldap, $base, "(modifyTimestamp<=" . $_SERVER["REQUEST_TIME"] . ")"); }
function safe10($ldap, $base, $flags, $name, $user) { return ldap_search($ldap, $base, "(port={$_SERVER['SERVER_PORT']})"); }
function safe11($ldap, $base, $flags, $name, $user) { return ldap_search($ldap, $base, "(argc=$_SERVER[argc])"); }
function safe12($ldap, $base, $flags, $name, $user) { return ldap_search($ldap, $base, `echo {$_GET['user']}`); }
