<?php

function componentBySpread(array $options): void
{
    echo parse_url("http:///broken", ...$options)["host"];
}
