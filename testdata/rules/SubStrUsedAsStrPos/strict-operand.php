<?php
// `===` never matches a non-string operand, while strpos() casts its needle:
// the fix needs an operand that is a string by syntax or by declared type.
class Route { public static function prefix() { return '/app'; } }
/** @param string $doc */
function prefixes(string $uri, $any, $doc, int $code, ?string $maybe, string $base, string $tail) {
    $r = [];
    $r[] = <weak_warning descr="Use 'strpos($uri, $any) === 0' instead.">substr($uri, 0, strlen($any)) === $any</weak_warning>;
    $r[] = <weak_warning descr="Use 'strpos($uri, $doc) === 0' instead.">substr($uri, 0, strlen($doc)) === $doc</weak_warning>;
    $r[] = <weak_warning descr="Use 'strpos($uri, $code) !== 0' instead.">substr($uri, 0, strlen($code)) !== $code</weak_warning>;
    $r[] = <weak_warning descr="Use 'strpos($uri, $maybe) === 0' instead.">substr($uri, 0, strlen($maybe)) === $maybe</weak_warning>;
    $r[] = <weak_warning descr="Use 'strpos($uri, $base) === 0' instead.">substr($uri, 0, strlen($base)) === $base</weak_warning>;
    $r[] = <weak_warning descr="Use 'strpos($uri, &quot;$base/$tail&quot;) !== 0' instead.">"$base/$tail" !== substr($uri, 0, strlen("$base/$tail"))</weak_warning>;
    $r[] = <weak_warning descr="Use 'mb_strpos($uri, $base . $tail) === 0' instead.">mb_substr($uri, 0, mb_strlen($base . $tail)) === $base . $tail</weak_warning>;
    // Written twice here, evaluated once after the rewrite.
    $r[] = <weak_warning descr="Use 'strpos($uri, Route::prefix()) === 0' instead.">substr($uri, 0, strlen(Route::prefix())) === Route::prefix()</weak_warning>;
    $r[] = <weak_warning descr="Use 'strpos($uri, &quot;{$any->name()}&quot;) === 0' instead.">substr($uri, 0, strlen("{$any->name()}")) === "{$any->name()}"</weak_warning>;
    return $r;
}
