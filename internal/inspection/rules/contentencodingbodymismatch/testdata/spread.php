<?php

function framedBySpread(array $options): void
{
    header("Content-Encoding: gzip");
    echo gzcompress("body", ...$options);
}
