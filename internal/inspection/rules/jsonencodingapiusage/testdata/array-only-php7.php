<?php

// Before PHP 8.0, PHP's own functions only warn about null and go on.
function keys(string $raw): array
{
    return array_keys(<weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($raw, true)</weak_warning>);
}
