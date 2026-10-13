<?php

function failureValueIsNotFalse($input, $ids, $flags)
{
    // A configured default replaces false on failure.
    if (!filter_var($input, FILTER_VALIDATE_INT, ['options' => ['default' => 0]])) {
        return 'a';
    }
    if (!filter_var($input, FILTER_VALIDATE_INT, ['options' => ['default' => null]])) {
        return 'b';
    }
    if (!filter_var($input, FILTER_VALIDATE_INT, ['options' => ['min_range' => 0, 'default' => 5]])) {
        return 'c';
    }
    if (!filter_var($input, FILTER_VALIDATE_INT, ['flags' => FILTER_FLAG_ALLOW_HEX, 'options' => ['default' => 2]])) {
        return 'd';
    }
    // Array-shaping flags make the result an array.
    if (!filter_var($ids, FILTER_VALIDATE_INT, FILTER_REQUIRE_ARRAY)) {
        return 'e';
    }
    if (!filter_var($ids, FILTER_VALIDATE_INT, FILTER_FORCE_ARRAY)) {
        return 'f';
    }
    if (!filter_var($ids, FILTER_VALIDATE_INT, FILTER_FLAG_ALLOW_OCTAL | FILTER_FORCE_ARRAY)) {
        return 'g';
    }
    if (!filter_var($ids, FILTER_VALIDATE_INT, ['flags' => FILTER_REQUIRE_ARRAY])) {
        return 'h';
    }
    if (!filter_var($ids, FILTER_VALIDATE_INT, ['flags' => FILTER_FLAG_ALLOW_HEX | FILTER_FORCE_ARRAY, 'options' => ['min_range' => -3]])) {
        return 'i';
    }
    // Unknown flags or options prove nothing.
    if (!filter_var($input, FILTER_VALIDATE_INT, $flags)) {
        return 'j';
    }
    if (!filter_var($input, FILTER_VALIDATE_INT, ['flags' => $flags])) {
        return 'k';
    }
    return 'ok';
}
