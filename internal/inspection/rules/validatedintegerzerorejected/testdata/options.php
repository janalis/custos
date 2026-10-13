<?php

function zeroStillRejected($input)
{
    if (!<warning descr="Compare integer validation failure strictly with false.">filter_var($input, FILTER_VALIDATE_INT, FILTER_FLAG_ALLOW_HEX)</warning>) {
        return 'a';
    }
    if (!<warning descr="Compare integer validation failure strictly with false.">filter_var($input, FILTER_VALIDATE_INT, FILTER_REQUIRE_SCALAR)</warning>) {
        return 'b';
    }
    if (!<warning descr="Compare integer validation failure strictly with false.">filter_var($input, FILTER_VALIDATE_INT, ['flags' => FILTER_FLAG_ALLOW_OCTAL])</warning>) {
        return 'c';
    }
    if (!<warning descr="Compare integer validation failure strictly with false.">filter_var($input, FILTER_VALIDATE_INT, ['options' => ['min_range' => -5, 'max_range' => 5]])</warning>) {
        return 'd';
    }
    if (!<warning descr="Compare integer validation failure strictly with false.">filter_var($input, FILTER_VALIDATE_INT, ['flags' => FILTER_FLAG_ALLOW_HEX, 'options' => ['min_range' => 0]])</warning>) {
        return 'e';
    }
    return 'ok';
}
