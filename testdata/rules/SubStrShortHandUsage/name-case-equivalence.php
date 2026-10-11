<?php
function tail($o, $from) {
    // Method calls are not known to be repeatable: reported without a fix.
    $a = substr($o->Text(), $from, <warning descr="The length 'strlen($o-&gt;text()) - $from' is unnecessary; remove it.">strlen($o->text()) - $from</warning>);
    $b = substr(Cache::$map['k'], $from, <warning descr="The length 'strlen(cache::$map['k']) - $from' is unnecessary; remove it.">strlen(cache::$map['k']) - $from</warning>);
    return [$a, $b];
}
