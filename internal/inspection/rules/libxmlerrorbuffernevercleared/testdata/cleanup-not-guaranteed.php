<?php

function returnsBeforeClear(array $docs): array
{
    libxml_use_internal_errors(true);
    <warning descr="Clear collected XML errors during repeated parsing.">foreach ($docs as $xml) {
        simplexml_load_string($xml);
    }</warning>
    $errors = libxml_get_errors();
    if ($errors === []) {
        return [];
    }
    libxml_clear_errors();

    return $errors;
}

function clearsConditionally(array $docs, bool $reset): void
{
    libxml_use_internal_errors(true);
    <warning descr="Clear collected XML errors during repeated parsing.">foreach ($docs as $xml) {
        simplexml_load_string($xml);
    }</warning>
    if ($reset) {
        libxml_clear_errors();
    }
}

function returnsThenLoops(array $docs, array $retries): array
{
    libxml_use_internal_errors(true);
    <warning descr="Clear collected XML errors during repeated parsing.">foreach ($docs as $xml) {
        simplexml_load_string($xml);
    }</warning>
    if ($docs === []) {
        return [];
    } else {
        foreach ($retries as $retry) {
            break;
        }
    }
    libxml_clear_errors();

    return $retries;
}
