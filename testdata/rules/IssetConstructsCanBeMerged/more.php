<?php
$a = $ok && isset($p) && <weak_warning descr="Merge this check into the preceding isset() call.">isset($q)</weak_warning>;
$b = isset($p) && $ok && <weak_warning descr="Merge this check into the preceding isset() call.">isset($q)</weak_warning> && isset($r);
if ((isset($p) && <weak_warning descr="Merge this check into the preceding isset() call.">isset($q)</weak_warning>)) {}
$c = isset($p) && ((<weak_warning descr="Merge this check into the preceding isset() call.">isset($q)</weak_warning> && isset($r)));
