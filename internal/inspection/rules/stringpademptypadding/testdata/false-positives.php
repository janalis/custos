<?php
str_pad('q', 1, '');

function shrunkWidth(): string
{
    $width = 3;
    $width--;
    $width--;
    return str_pad('ab', $width, '');
}

function padChosenLater(): string
{
    $pad = '';
    $choose = function () use (&$pad) {
        $pad = '*';
    };
    $choose();
    return str_pad('ab', 5, $pad);
}

function widthCopied(): string
{
    $width = 3;
    $width--;
    $width--;
    $copy = $width;
    return str_pad('ab', $copy, '');
}

function padExtracted(array $settings): string
{
    $pad = '';
    extract($settings);
    return str_pad('ab', 5, $pad);
}
