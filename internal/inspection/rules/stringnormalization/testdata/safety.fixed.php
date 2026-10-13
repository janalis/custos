<?php
function safety($code, $mode) {
    $r = [];
    // cutting around a multibyte conversion: reported, never rewritten
    $r[] = mb_substr(mb_strtoupper($code), 0, 6);
    $r[] = substr(mb_convert_case($code, MB_CASE_UPPER), 1);
    $r[] = mb_substr(mb_convert_case($code, \MB_CASE_FOLD_SIMPLE), 1);
    // title case or an unknown mode depends on the position: not cut first
    $r[] = substr(mb_convert_case($code, MB_CASE_TITLE), 1);
    $r[] = mb_substr(mb_convert_case($code, $mode), 1);
    $r[] = substr(mb_convert_case($code), 1);
    // byte-wise conversions keep every offset
    $r[] = strtoupper(mb_substr($code, 0, 6));
    // trimming is still swapped
    $r[] = mb_convert_case(ltrim($code, '-'), MB_CASE_TITLE);
    // a repeated call with another mode, encoding or delimiter is not redundant
    $r[] = mb_convert_case(mb_convert_case($code, MB_CASE_UPPER), MB_CASE_LOWER);
    $r[] = mb_strtolower(mb_strtolower($code, 'ISO-8859-1'));
    $r[] = mb_strtolower(mb_strtolower($code), 'ISO-8859-1');
    $r[] = ucwords(ucwords($code), '-');
    $r[] = mb_convert_case($code, MB_CASE_UPPER);
    $r[] = mb_strtolower($code, 'UTF-8');
    // octal escapes can spell letters
    $r[] = trim(strtolower($code), "\101");
    $r[] = trim(strtolower($code), "-\\\60");
    $r[] = trim(strtolower($code), <<<EOT
    \101
    EOT);
    $r[] = strtolower(trim($code, "\\101"));
    $r[] = strtolower(trim($code, '\101'));
    $r[] = strtolower(trim($code, <<<'EOT'
    \101
    EOT));
    return $r;
}
