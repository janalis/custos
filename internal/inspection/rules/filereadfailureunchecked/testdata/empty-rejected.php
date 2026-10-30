<?php
// PHP 8 throws a ValueError when these parse empty text.
final class Loader
{
    public function loadXML(string $source): bool { return $source !== ''; }
}

function parse(DOMDocument $dom, XMLReader $reader, Loader $own, $any, string $path): array
{
    return [
        $dom->loadXML(file_get_contents($path)),
        $dom->schemaValidateSource(file_get_contents($path)),
        $reader->XML(file_get_contents($path)),
        <warning descr="Handle file reading failure before using its result.">$own->loadXML(file_get_contents($path))</warning>,
        $any->loadXML(file_get_contents($path)),
        <warning descr="Handle file reading failure before using its result.">$dom->createAttributeNS('urn:a', file_get_contents($path))</warning>,
    ];
}
