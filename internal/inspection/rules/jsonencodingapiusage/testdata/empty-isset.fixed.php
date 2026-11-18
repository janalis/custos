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
    $json = json_encode($row, JSON_THROW_ON_ERROR);
    return isset($json) ? $json : '';
}

function readTags(string $out): array
{
    $documents = json_decode($out, true);
    return $documents[0]['tags'] ?? [];
}

function readTitle(string $out): string
{
    $doc = json_decode($out);
    return isset($doc->title) ? $doc->title : '';
}

function readFirst(string $out): array
{
    $doc = json_decode($out, true, 512, JSON_THROW_ON_ERROR);
    return $doc[$doc['first'] ?? 0];
}

function readCount(string $out): int
{
    $doc = json_decode($out, true, 512, JSON_THROW_ON_ERROR);
    return count($doc['items']);
}

function readLength(string $out): int
{
    $doc = json_decode($out, false, 512, JSON_THROW_ON_ERROR);
    return $doc->items->length;
}

function lookups(string $out, array $map, object $o): array
{
    $key = json_decode($out, false, 512, JSON_THROW_ON_ERROR);
    $name = json_decode($out, false, 512, JSON_THROW_ON_ERROR);
    $doc = json_decode($out, true, 512, JSON_THROW_ON_ERROR);
    return [$map[$key], $o->{$name}, $doc['a'] + 1];
}
