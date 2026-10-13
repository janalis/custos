<?php
function wrong_arm($u) { $i = posix_getpwnam($u); return $i ? null : <warning descr="Reject account lookup failure before indexing.">$i['dir']</warning>; }
function wrong_branch($u) { $i = posix_getpwnam($u); if (!$i) { echo <warning descr="Reject account lookup failure before indexing.">$i['dir']</warning>; } }
function plain_isset($u) { $i = posix_getpwnam($u); if (isset($i)) { echo <warning descr="Reject account lookup failure before indexing.">$i['dir']</warning>; } }
function null_check($u) { $i = posix_getpwnam($u); if ($i !== null) { echo <warning descr="Reject account lookup failure before indexing.">$i['dir']</warning>; } }
function unrelated($u, $o) { $i = posix_getpwnam($u); if (is_array($o)) { echo <warning descr="Reject account lookup failure before indexing.">$i['dir']</warning>; } }
function no_exit($u) { $i = posix_getpwnam($u); if (!$i) { echo 0; } echo <warning descr="Reject account lookup failure before indexing.">$i['dir']</warning>; }
function keeps_failure($u) { $i = posix_getpwnam($u); return $i || <warning descr="Reject account lookup failure before indexing.">$i['dir']</warning>; }
function eager($u) { $i = posix_getpwnam($u); return !$i . <warning descr="Reject account lookup failure before indexing.">$i['dir']</warning>; }
function no_abort($u) { $i = posix_getpwnam($u); $i or print('x'); return <warning descr="Reject account lookup failure before indexing.">$i['dir']</warning>; }
function unrelated_else($u, $o) { $i = posix_getpwnam($u); echo $o; if ($o) { echo 0; } else { echo <warning descr="Reject account lookup failure before indexing.">$i['dir']</warning>; } }
