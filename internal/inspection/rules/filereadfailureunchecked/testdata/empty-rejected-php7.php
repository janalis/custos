<?php
// Before PHP 8 empty text only warns.
final class Loader
{
    public function loadXML(string $source): bool { return $source !== ''; }
}

function parse(DOMDocument $dom, XMLReader $reader, Loader $own, $any, string $path): array
{
    return [
        <warning descr="Handle file reading failure before using its result.">$dom->loadXML(file_get_contents($path))</warning>,
        <warning descr="Handle file reading failure before using its result.">$dom->schemaValidateSource(file_get_contents($path))</warning>,
        <warning descr="Handle file reading failure before using its result.">$reader->XML(file_get_contents($path))</warning>,
        <warning descr="Handle file reading failure before using its result.">$own->loadXML(file_get_contents($path))</warning>,
    ];
}
