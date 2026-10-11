<?php
function truthiness(string $rules, string $text): int { $s = transliterator_transliterate($rules, $text); if (!$s) { return 0; } return <warning descr="Check transliteration before consuming text.">strlen($s)</warning>; }
function looseComparison(string $rules, string $text): int { $s = transliterator_transliterate($rules, $text); return $s == false ? 0 : <warning descr="Check transliteration before consuming text.">strlen($s)</warning>; }
function wrongBranch(string $rules, string $text): int { $s = transliterator_transliterate($rules, $text); return $s === false ? <warning descr="Check transliteration before consuming text.">strlen($s)</warning> : 0; }
function reboundAfterCheck(string $rules, string $text): int { $s = transliterator_transliterate($rules, $text); if ($s === false) { return 0; } $s = transliterator_transliterate($rules, $text); return <warning descr="Check transliteration before consuming text.">strlen($s)</warning>; }
function otherVariable(string $rules, string $text, $other): bool { $s = transliterator_transliterate($rules, $text); return $other !== false && <warning descr="Check transliteration before consuming text.">strlen($s)</warning> > 3; }
