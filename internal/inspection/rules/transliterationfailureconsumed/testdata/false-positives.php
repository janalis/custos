<?php
function ternaryGuard(string $rules, string $text): int { $s = transliterator_transliterate($rules, $text); return $s === false ? 0 : strlen($s); }
function conjunction(string $rules, string $text): bool { $s = transliterator_transliterate($rules, $text); return $s !== false && strlen($s) > 3; }
function disjunction(string $rules, string $text): bool { $s = transliterator_transliterate($rules, $text); return false === $s || strlen($s) > 3; }
function assignedInCondition(string $rules, string $text): int { if (($s = transliterator_transliterate($rules, $text)) === false) { return 0; } return strlen($s); }
function typeChecked(string $rules, string $text): int { $s = transliterator_transliterate($rules, $text); if (!is_string($s)) { return 0; } return strlen($s); }
function typeBranch(Transliterator $tr, string $text): string { $s = $tr->transliterate($text); if (is_string($s)) { return strtoupper($s); } return ''; }
function matchArm(string $rules, string $text): int { $s = transliterator_transliterate($rules, $text); return match (true) { $s === false => 0, default => strlen($s) }; }
function elseBranch(Transliterator $tr, string $text): string { $s = $tr->transliterate($text); if ($s === false) { return ''; } else { return strtoupper($s); } }
function emptyOrFailed(string $rules, string $text): int { $s = transliterator_transliterate($rules, $text); if ($s === false || $s === '') { return 0; } return strlen($s); }
