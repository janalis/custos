<?php
// A case-insensitive search only matches when the compared value is already folded.
function verbs($uri, $verb) {
    $r = [];
    $r[] = strtoupper(substr($uri, 0, strlen($verb))) === $verb;
    $r[] = strtoupper(substr($uri, 0, 3)) === 'get';
    $r[] = strtolower(substr($uri, 0, 3)) === 'Get';
    $r[] = mb_strtoupper(mb_substr($uri, 0, 1)) === 'É';
    return $r;
}
