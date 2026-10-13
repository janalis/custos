<?php
function tail($o, $from) {
    // Method calls are not known to be repeatable: reported without a fix.
    $a = substr($o->Text(), $from, strlen($o->text()) - $from);
    $b = substr(Cache::$map['k'], $from);
    return [$a, $b];
}
