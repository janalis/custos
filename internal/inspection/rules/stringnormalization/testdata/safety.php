<?php
function safety($code, $mode) {
    $r = [];
    // cutting around a multibyte conversion: reported, never rewritten
    $r[] = <weak_warning descr="Cut first, then change the case (check multibyte input by hand).">mb_substr(mb_strtoupper($code), 0, 6)</weak_warning>;
    $r[] = <weak_warning descr="Cut first, then change the case (check multibyte input by hand).">substr(mb_convert_case($code, MB_CASE_UPPER), 1)</weak_warning>;
    $r[] = <weak_warning descr="Cut first, then change the case (check multibyte input by hand).">mb_substr(mb_convert_case($code, \MB_CASE_FOLD_SIMPLE), 1)</weak_warning>;
    // title case or an unknown mode depends on the position: not cut first
    $r[] = substr(mb_convert_case($code, MB_CASE_TITLE), 1);
    $r[] = mb_substr(mb_convert_case($code, $mode), 1);
    $r[] = substr(mb_convert_case($code), 1);
    // byte-wise conversions keep every offset
    $r[] = <weak_warning descr="Cut first, then change the case: 'strtoupper(mb_substr($code, 0, 6))'.">mb_substr(strtoupper($code), 0, 6)</weak_warning>;
    // trimming is still swapped
    $r[] = <weak_warning descr="Cut first, then change the case: 'mb_convert_case(ltrim($code, '-'), MB_CASE_TITLE)'.">ltrim(mb_convert_case($code, MB_CASE_TITLE), '-')</weak_warning>;
    // a repeated call with another mode, encoding or delimiter is not redundant
    $r[] = mb_convert_case(mb_convert_case($code, MB_CASE_UPPER), MB_CASE_LOWER);
    $r[] = mb_strtolower(mb_strtolower($code, 'ISO-8859-1'));
    $r[] = mb_strtolower(mb_strtolower($code), 'ISO-8859-1');
    $r[] = ucwords(ucwords($code), '-');
    $r[] = mb_convert_case(<weak_warning descr="The inner 'mb_convert_case(...)' call has no effect here.">mb_convert_case($code, MB_CASE_UPPER)</weak_warning>, MB_CASE_UPPER);
    $r[] = mb_strtolower(<weak_warning descr="The inner 'mb_strtolower(...)' call has no effect here.">mb_strtolower($code, 'UTF-8')</weak_warning>, 'UTF-8');
    // octal escapes can spell letters
    $r[] = trim(strtolower($code), "\101");
    $r[] = trim(strtolower($code), "-\\\60");
    $r[] = trim(strtolower($code), <<<EOT
    \101
    EOT);
    $r[] = <weak_warning descr="Cut first, then change the case: 'strtolower(trim($code, &quot;\\101&quot;))'.">trim(strtolower($code), "\\101")</weak_warning>;
    $r[] = <weak_warning descr="Cut first, then change the case: 'strtolower(trim($code, '\101'))'.">trim(strtolower($code), '\101')</weak_warning>;
    $r[] = <weak_warning descr="Cut first, then change the case: 'strtolower(trim($code, &lt;&lt;&lt;'EOT'
    \101
    EOT))'.">trim(strtolower($code), <<<'EOT'
    \101
    EOT)</weak_warning>;
    return $r;
}
