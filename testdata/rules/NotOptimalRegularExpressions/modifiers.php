<?php
function modifiersDemo($line, $m) {
    preg_match('%k\d+%', $line, $m);
    preg_match('[k\d+]x', $line, $m);
    preg_match('', $line, $m);
    preg_match(<warning descr="Pattern has no valid delimiters.">'k\d+'</warning>, $line, $m);
    preg_match(<warning descr="Pattern has no valid delimiters.">')k\d('</warning>, $line, $m);
    preg_match(<warning descr="Pattern has no valid delimiters.">'/k\d/9'</warning>, $line, $m);
    preg_quote('k.d', '/');
    preg_quote('/k/Q', '/');
    preg_replace(<error descr="The /e flag was removed from PCRE; use a callback replacement.">'#k(\d)#e'</error>, '$1', $line);
    preg_replace([<error descr="The /e flag was removed from PCRE; use a callback replacement.">'#k(\d)#e'</error>], '$1', $line);
    preg_match(<error descr="'Q' is not a valid PCRE modifier.">'/k\d/Q'</error>, $line, $m);
    preg_match(<error descr="'Q' is not a valid PCRE modifier."><error descr="'Q' is not a valid PCRE modifier.">'/k\d/QQ'</error></error>, $line, $m);
    preg_match(<error descr="'n' is not a valid PCRE modifier.">'/(k)\d/n'</error>, $line, $m);

    preg_match('/^k\d$/D', $line, $m);
    preg_match(<weak_warning descr="The /D flag is pointless: the pattern has no '$'.">'/^k\d/D'</weak_warning>, $line, $m);
    preg_match(<weak_warning descr="The /D flag is pointless: the pattern has no '$'.">'/k\$/D'</weak_warning>, $line, $m);
    preg_match(<weak_warning descr="The /D flag has no effect together with /m.">'/k\d$/mD'</weak_warning>, $line, $m);

    preg_match('/k.\d/s', $line, $m);
    preg_match('/k\[.\]/s', $line, $m);
    preg_match(<weak_warning descr="The /s flag is pointless: the pattern has no '.'.">'/k\d+/s'</weak_warning>, $line, $m);
    preg_match(<weak_warning descr="The /s flag is pointless: the pattern has no '.'.">'/k[.,]\d/s'</weak_warning>, $line, $m);
    preg_match(<weak_warning descr="The /s flag is pointless: the pattern has no '.'.">'/k\.\d/s'</weak_warning>, $line, $m);

    preg_match('/k\d/i', $line, $m);
    preg_match(<weak_warning descr="The /i flag is pointless: the pattern has no letters.">'/\\d/i'</weak_warning>, $line, $m); // the PHP value is /\d/i
    preg_match('/\\\\d/i', $line, $m); // an escaped backslash, then the letter d
    preg_match("/^[a-z]*\$/D", $line);
    preg_match("/^[0-5]*\x24/D", $line);
    preg_match("/^\x2e/s", $line);
    preg_match('/ид/iu', $line, $m);
    preg_match(<weak_warning descr="The /i flag is pointless: the pattern has no letters.">'/\d{3}-\d{2}/i'</weak_warning>, $line, $m);
    preg_match(<weak_warning descr="The /i flag is pointless: the pattern has no letters.">'/
        \s+ \W
    /ix'</weak_warning>, $line, $m);

    $rx = <warning descr="Pattern has no valid delimiters.">'k\w+'</warning>;
    preg_match($rx, $line, $m);
    preg_match(<error descr="The /r flag needs /u to take effect."><error descr="The /r flag needs /i to take effect."><error descr="'r' is not a valid PCRE modifier.">'/kelvin/r'</error></error></error>, $line, $m);
}
function decodedEscapes(string $class)
{
    // The pattern is ^PhpOffice\\PhpSpreadsheet\\ (escaped backslashes), not \P.
    return preg_match('/^PhpOffice\\\PhpSpreadsheet\\\/', $class);
}
