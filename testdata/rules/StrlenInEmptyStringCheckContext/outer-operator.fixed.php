<?php
function outer($s, string $t, $u, bool $on)
{
    $r = [];
    $r[] = !((string)$s === '');
    $r[] = (int) ($t === '');
    $r[] = ((string)$s === '') . 'x';
    $r[] = ((string)$s === '') < 1;
    $r[] = ((string)$s === '') === ((string)$u === '');
    $r[] = ($t !== '') === $on;
    $r[] = $on == ($t === '');
    $r[] = -((string)$s === '') + 1;
    $r[] = ((string)$s === '') <=> $on;
    $r[] = (string)$s !== '' & $on;
    $r[] = (string)$s === '' ?? $on;
    $r[] = $on xor (string)$s === '';
    return $r;
}
echo json_encode([outer('abc', '', 'z', true), outer('', 'q', '', false)]), "\n";
