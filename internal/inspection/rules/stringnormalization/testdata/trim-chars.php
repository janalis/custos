<?php
function chars($code, $sep) {
    $r = [];
    $r[] = trim(strtolower($code), '-', 'x');               // too many arguments
    $r[] = trim(strtolower($code), 5);                      // not a string
    $r[] = trim(strtolower($code), "-$sep");                // interpolated: unknown characters
    $r[] = trim(strtolower($code), `ls`);                   // shell command
    $r[] = trim(strtolower($code), "-
");                                                         // multi-line quoted literal
    $r[] = trim(strtolower($code), b'abc');                 // letters behind a binary prefix
    $r[] = trim(strtolower($code), <<<EOT
    xyz
    EOT);                                                   // heredoc with letters
    $r[] = trim(strtolower($code), '@..Z');                 // range covering letters
    $r[] = trim(strtolower($code), '!..~');
    $r[] = <weak_warning descr="Cut first, then change the case: 'strtolower(trim($code, '0..9'))'.">trim(strtolower($code), '0..9')</weak_warning>;
    $r[] = <weak_warning descr="Cut first, then change the case: 'strtolower(trim($code, b'-'))'.">trim(strtolower($code), b'-')</weak_warning>;
    $r[] = <weak_warning descr="Cut first, then change the case: 'strtolower(trim($code, &lt;&lt;&lt;'EOT'
    -/
    EOT))'.">trim(strtolower($code), <<<'EOT'
    -/
    EOT)</weak_warning>;
    return $r;
}
