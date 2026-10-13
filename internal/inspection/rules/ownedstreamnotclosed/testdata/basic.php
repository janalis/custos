<?php
function bad($path) { $h = fopen($path, 'rb'); if ($h === false) { return false; } <warning descr="Close the owned stream before leaving its scope.">return fgets($h);</warning> }
function lent($path) { $h = fopen($path, 'rb'); if ($h === false) { return 0; } <warning descr="Close the owned stream before leaving its scope.">return strlen(fgets($h)) + ftell($h);</warning> }
function partlyClosed($path, $skip) {
    $h = fopen($path, 'rb');
    if (!$h || $h === false) {
        return false;
    }
    if ($skip) {
        fclose($h);
        return '';
    }
    $unrelated = function () use ($skip) { return $skip; };
    <warning descr="Close the owned stream before leaving its scope.">return fread(($h), 16);</warning>
}
