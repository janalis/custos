<?php
function strict_ternary($u) { $i = posix_getpwnam($u); return $i !== false ? $i['dir'] : null; }
function strict_else_arm($u) { $i = posix_getpwnam($u); return $i === false ? null : $i['dir']; }
function strict_short_circuit($u) { $i = posix_getpwnam($u); if ($i === false || $i['uid'] === 0) { return null; } return $i['dir']; }
function falsy_return($u) { $i = posix_getpwnam($u); if (!$i) { return null; } return $i['dir']; }
function truthy_ternary($u) { $i = posix_getpwnam($u); return $i ? $i['dir'] : null; }
function array_branch($u) { $i = posix_getpwnam($u); if (is_array($i)) { return $i['dir']; } return null; }
function isset_then_read($u) { $i = posix_getpwnam($u); return isset($i['dir']) && $i['dir'] !== ''; }
function not_empty_key($u) { $i = posix_getpwuid($u); if (!empty($i['dir'])) { echo $i['dir']; } }
function loose_exit($u) { $i = posix_getpwuid($u); if ($i == false) { return; } echo $i['dir']; }
function else_branch($u) { $i = posix_getpwuid($u); if (!$i) { echo 0; } else { echo $i['dir']; } }
function elseif_branch($u, $o) { $i = posix_getpwuid($u); if ($o) { echo 0; } elseif (!$i) { echo 1; } elseif ($o > 1) { echo $i['uid']; } else { echo $i['dir']; } }
function or_die($u) { $i = posix_getpwnam($u) or die('unknown'); return $i['dir']; }
function and_exit($u) { $i = posix_getpwnam($u); $i === false and exit(1); return $i['dir']; }
function assigned_exit($u) { if (!($i = posix_getpwnam($u))) { return null; } return $i['dir']; }
