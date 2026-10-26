<?php
function sent(string $json, bool $b, ?bool $nb, string $s, int|string $is) {
    $content = json_decode($json, true);
    if ('true' != $content['success']) {}
    if ($b == 'TRUE') {}
    if ($nb == 'true') {}
    if ('true' == $unknown) {}
    if (<warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">$s == 'true'</warning>) {}
    if (<warning descr="Use '!==' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">'false' != $content['ok']</warning>) {}
    if (<warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">$is == 'true'</warning>) {}
}
