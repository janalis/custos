<?php

function encodedBySpread(array $options): string
{
    $context = inflate_init(ZLIB_ENCODING_RAW);

    return inflate_add($context, gzencode("body", ...$options));
}
