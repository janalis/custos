<?php
class Suffixes {
    const Array = 'abc';
    const ARRAY = 'xy';
    const List = '.xy';
    static function pick($array = 'abc', $ARRAY = 'xy') { return $array . $ARRAY; }
}
function endsWithKeywordNamed(string $h): array {
    return [
        // Two different constants: constant names are case-sensitive.
        substr($h, -strlen(Suffixes::Array)) === Suffixes::ARRAY,
        // Two different named arguments.
        substr($h, -strlen(Suffixes::pick(array: 'a'))) === Suffixes::pick(ARRAY: 'a'),
        str_ends_with($h, suffixes::List),
        substr($h, -strlen(Suffixes::Pick(array: 'a'))) === suffixes::pick(array: 'a'),
        substr($h, -strlen(Suffixes::class)) === Suffixes::CLASS,
        substr($h, -strlen(Suffixes::{'List'})) === suffixes::{'List'},
    ];
}
