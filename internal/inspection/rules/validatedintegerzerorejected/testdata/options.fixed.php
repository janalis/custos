<?php

function zeroStillRejected($input)
{
    if ((filter_var($input, FILTER_VALIDATE_INT, FILTER_FLAG_ALLOW_HEX) === false)) {
        return 'a';
    }
    if ((filter_var($input, FILTER_VALIDATE_INT, FILTER_REQUIRE_SCALAR) === false)) {
        return 'b';
    }
    if ((filter_var($input, FILTER_VALIDATE_INT, ['flags' => FILTER_FLAG_ALLOW_OCTAL]) === false)) {
        return 'c';
    }
    if ((filter_var($input, FILTER_VALIDATE_INT, ['options' => ['min_range' => -5, 'max_range' => 5]]) === false)) {
        return 'd';
    }
    if ((filter_var($input, FILTER_VALIDATE_INT, ['flags' => FILTER_FLAG_ALLOW_HEX, 'options' => ['min_range' => 0]]) === false)) {
        return 'e';
    }
    return 'ok';
}
