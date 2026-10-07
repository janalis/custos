<?php
function feeds(string $dir, string $cls, int $flags, string $nsUri) {
    $empty = simplexml_load_file();
    $a = simplexml_load_string(file_get_contents($dir . '/rss.xml'));
    $b = simplexml_load_string(file_get_contents($dir . '/atom.xml'), $cls);
    $c = simplexml_load_string(file_get_contents(getenv('FEED')), $cls, $flags, $nsUri, true);
    $d = $this->simplexml_load_file('x.xml');
    return [$a, $b, $c, $d, $empty];
}
