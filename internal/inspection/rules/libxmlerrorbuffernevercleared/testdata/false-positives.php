<?php

function collectThenClear(array $docs): array
{
    libxml_use_internal_errors(true);
    $d = new DOMDocument();
    foreach ($docs as $xml) {
        $d->loadXML($xml);
    }
    $errors = libxml_get_errors();
    libxml_clear_errors();

    return $errors;
}

function printThenClear(array $docs): void
{
    libxml_use_internal_errors(true);
    foreach ($docs as $xml) {
        simplexml_load_string($xml);
    }
    foreach (libxml_get_errors() as $error) {
        if ($error->level < LIBXML_ERR_ERROR) {
            continue;
        }
        echo trim($error->message), "\n";
    }
    libxml_clear_errors();
}
