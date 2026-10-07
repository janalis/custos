<?php
$title = getTitle() ? getTitle() : 'none';
$next  = next($it) ? next($it) : null;
$row   = $stmt->fetch() ? $stmt->fetch() : [];
$one   = Repo::find(1) ? Repo::find(1) : null;
$obj   = new Box() ? new Box() : null;
$inc   = $i++ ? $i++ : 0;
$set   = ($x = load()) ? ($x = load()) : 0;
$key   = $map[key($map)] ? $map[key($map)] : 0;
$fine  = <weak_warning descr="Use the short ternary: '$map[$k] ?: fallback()'.">$map[$k] ? $map[$k] : fallback()</weak_warning>;
