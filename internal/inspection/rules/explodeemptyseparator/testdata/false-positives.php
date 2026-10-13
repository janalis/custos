<?php
// The separator variable may be rewritten without a plain assignment between
// its empty literal and the call: no definite empty separator.
function viaReferenceCapture(string $line) {
    $glue = '';
    $pick = function () use (&$glue) { $glue = '|'; };
    $pick();
    return explode($glue, $line);
}
function viaGlobal(string $line) {
    global $fieldGlue;
    $fieldGlue = '';
    loadSettings();
    return explode($fieldGlue, $line);
}
function loadSettings() { global $fieldGlue; $fieldGlue = ';'; }
function viaExtract(string $line, array $settings) {
    $glue = '';
    \extract($settings);
    return explode($glue, $line);
}
function viaQueryImport(string $line, string $query) {
    $glue = '';
    parse_str($query);
    return explode($glue, $line);
}
function viaIncrement(string $line) {
    $glue = '';
    $glue++;
    return explode(($glue), $line);
}
function viaReferencedItem(string $line) {
    $glue = '';
    $slots = array('glue' => &$glue);
    fillSlots($slots);
    return explode($glue, $line);
}
function fillSlots(array $slots) { $slots['glue'] = ':'; }
function viaCopy(string $line) {
    $first = '';
    $set = function () use (&$first) { $first = '/'; };
    $set();
    $second = $first;
    $glue = $second;
    return explode($glue, $line);
}
