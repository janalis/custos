<?php
function docPatterns($subject, $dynamic) {
    preg_match(<warning descr="Pattern has no valid delimiters."><<<'BROKEN'
not-delimited
BROKEN</warning>, $subject);
    preg_match(<warning descr="Pattern has no valid delimiters."><<<BROKEN
  not-delimited
  BROKEN</warning>, $subject);
    preg_match(<<<'VALID'
/record[a-f]+/
VALID, $subject);
    preg_match(<<<DYNAMIC
/$dynamic/
DYNAMIC, $subject);
    preg_match(<<<'EMPTY'
EMPTY, $subject);
}
