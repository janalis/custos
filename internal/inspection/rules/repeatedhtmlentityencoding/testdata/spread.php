<?php

function innerFlagsBySpread(array $options): string
{
    return htmlspecialchars(htmlspecialchars("&", ...$options));
}

function outerFlagsBySpread(array $options): string
{
    return htmlspecialchars(htmlspecialchars("&"), ...$options);
}
