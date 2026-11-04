<?php
function arrayDocPatterns($subject) {
    $constant = <<<'PATTERN'
/[a-f]+/
PATTERN;
    strlen($constant);
    preg_replace([$constant], ['x'], $subject);
}
