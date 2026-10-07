<?php

function forward(array $parts): string
{
    return implode(...$parts);
}
