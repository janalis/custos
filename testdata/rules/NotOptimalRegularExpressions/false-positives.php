<?php
function reviewClean(string $s, array $rest)
{
    // PHP skips spaces and line breaks among the modifiers
    preg_match("/abc/i\n", $s, $m);
    preg_match('/abc/i ', $s, $m);
    preg_match('/abc/i
', $s, $m);
    preg_match("\x2fabc\x2f", $s, $m);
    // an escaped bracket is text, not a class
    preg_match('/\[0-9]/', $s, $m);
    preg_match('/\\[0-9]/', $s, $m);
    preg_match('/x\[ab][ab]/', $s, $m);
    // the first argument is not the pattern
    preg_match(subject: 'tmp', pattern: $s);
    preg_match(subject: 'tmp', matches: $m);
    preg_match(...$rest);
}
