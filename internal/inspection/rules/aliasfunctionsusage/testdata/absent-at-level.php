<?php
if (!function_exists('is_real')) {
    function is_real($value) { return is_float($value); }
}

<warning descr="Use 'closedir(...)' instead of the alias 'close(...)'.">close</warning>($directory);
$r = <warning descr="Use 'recode_string(...)' instead of the alias 'recode(...)'.">recode</warning>('latin1..utf8', $text);
$f = is_real($ratio);
$n = <warning descr="Use 'count(...)' instead of the alias 'sizeof(...)'.">sizeof</warning>($rows);
