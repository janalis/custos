<?php

function readComposer(string $path): array
{
    $errors = [];
    $composer = json_decode(file_get_contents($path), false);
    if (empty($composer) || !is_object($composer)) {
        $errors[] = 'Parse error in composer.json';
    }
    return $errors;
}

function readSettings(string $raw): array
{
    $settings = json_decode($raw, true);
    return isset($settings) ? $settings : [];
}

function encodeRow(array $row): string
{
    $json = json_encode($row);
    return empty($json) ? '' : $json;
}

function encodeIsset(array $row): string
{
    $json = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($row)</weak_warning>;
    return isset($json) ? $json : '';
}
