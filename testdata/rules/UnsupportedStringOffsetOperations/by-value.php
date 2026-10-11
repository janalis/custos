<?php
function describe(string $title, string $suffix) {
    $length = strlen($title);
    $copy = $title;
    $parts = [$title, &$suffix];
    announce(text: $title, level: $length);
    announce($title, $length, $copy);
    shout(loud: $title);
    mystery(label: $title);
    <error descr="Appending with [] is not supported on strings (fatal error).">$title[]</error> = 'x';
}

function announce(string $text, int $level) {
}

function shout(string &$other, string $loud = '') {
}
