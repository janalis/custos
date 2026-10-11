<?php
ldap_search($ldap,$base,<warning descr="Escape request data for LDAP filter syntax.">"(uid=".$_GET["user"].")"</warning>);
// Interpolation, escaping for the wrong context and client-influenced server keys.
function case1($ldap, $base) { return ldap_search($ldap, $base, <warning descr="Escape request data for LDAP filter syntax.">"(uid={$_GET['user']})"</warning>); }
function case2($ldap, $base) { return ldap_list($ldap, $base, <warning descr="Escape request data for LDAP filter syntax.">"(uid=$_POST[user])"</warning>); }
function case3($ldap, $base) { return ldap_read($ldap, $base, <warning descr="Escape request data for LDAP filter syntax.">"(uid=$_COOKIE[0])"</warning>); }
function case4($ldap, $base) { return ldap_search($ldap, $base, <warning descr="Escape request data for LDAP filter syntax.">"(uid=" . ldap_escape($_GET["user"], "", LDAP_ESCAPE_DN) . ")"</warning>); }
function case5($ldap, $base) { return ldap_list($ldap, $base, <warning descr="Escape request data for LDAP filter syntax.">"(uid=" . ldap_escape($_GET["user"], "", 2) . ")"</warning>); }
function case6($ldap, $base) { return ldap_read($ldap, $base, <warning descr="Escape request data for LDAP filter syntax.">"(uid=" . ldap_escape($_GET["user"], "", LDAP_ESCAPE_DN | 4) . ")"</warning>); }
function case7($ldap, $base) { return ldap_search($ldap, $base, <warning descr="Escape request data for LDAP filter syntax.">"(uid=" . ldap_escape(flags: LDAP_ESCAPE_DN, value: $_GET["user"]) . ")"</warning>); }
function case8($ldap, $base) { return ldap_list($ldap, $base, <warning descr="Escape request data for LDAP filter syntax.">"(host=" . $_SERVER["HTTP_HOST"] . ")"</warning>); }
function case9($ldap, $base) { return ldap_read($ldap, $base, <warning descr="Escape request data for LDAP filter syntax.">"(host={$_SERVER['SERVER_NAME']})"</warning>); }
