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
    $doc = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($out, true)</weak_warning>;
    return $doc[$doc['first'] ?? 0];
}

function readCount(string $out): int
{
    $doc = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($out, true)</weak_warning>;
    return count($doc['items']);
}

function readLength(string $out): int
{
    $doc = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($out)</weak_warning>;
    return $doc->items->length;
}

function lookups(string $out, array $map, object $o): array
{
    $key = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($out)</weak_warning>;
    $name = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($out)</weak_warning>;
    $doc = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($out, true)</weak_warning>;
    return [$map[$key], $o->{$name}, $doc['a'] + 1];
}
