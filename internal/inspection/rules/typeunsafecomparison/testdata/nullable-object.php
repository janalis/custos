<?php
class Invoice {}

function check(?Invoice $n, Invoice $i, Invoice|null $m) {
    return [
        $n == '',
        '' != $m,
        ($n) <> "",
        <error descr="\Invoice has no __toString(), so it cannot be compared to a string.">$n == 'paid'</error>,
        <error descr="\Invoice has no __toString(), so it cannot be compared to a string.">$i == ''</error>,
    ];
}
