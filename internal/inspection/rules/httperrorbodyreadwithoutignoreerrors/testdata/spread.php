<?php

function optionsBySpread(array $options): string
{
    $context = stream_context_create(...$options);

    return file_get_contents("https://example.com/", false, $context);
}

function paramsBySpread(array $params): string
{
    $context = stream_context_create(["http" => ["method" => "GET"]], ...$params);

    return file_get_contents("https://example.com/", false, $context);
}
