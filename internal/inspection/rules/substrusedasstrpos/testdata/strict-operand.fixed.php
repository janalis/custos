<?php
// `===` never matches a non-string operand, while strpos() casts its needle:
// the fix needs an operand that is a string by syntax or by declared type.
class Route { public static function prefix() { return '/app'; } }
/** @param string $doc */
function prefixes(string $uri, $any, $doc, int $code, ?string $maybe, string $base, string $tail) {
    $r = [];
    $r[] = substr($uri, 0, strlen($any)) === $any;
    $r[] = substr($uri, 0, strlen($doc)) === $doc;
    $r[] = substr($uri, 0, strlen($code)) !== $code;
    $r[] = substr($uri, 0, strlen($maybe)) === $maybe;
    $r[] = strpos($uri, $base) === 0;
    $r[] = strpos($uri, "$base/$tail") !== 0;
    $r[] = mb_strpos($uri, $base . $tail) === 0;
    // Written twice here, evaluated once after the rewrite.
    $r[] = substr($uri, 0, strlen(Route::prefix())) === Route::prefix();
    $r[] = substr($uri, 0, strlen("{$any->name()}")) === "{$any->name()}";
    return $r;
}
