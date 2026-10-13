<?php
function feeds(string $dir, string $cls, int $flags, string $nsUri) {
    $empty = simplexml_load_file();
    $a = <error descr="simplexml_load_file() is affected by PHP bug #62577; load the contents with file_get_contents() and parse them with simplexml_load_string().">simplexml_load_file($dir . '/rss.xml')</error>;
    $b = <error descr="simplexml_load_file() is affected by PHP bug #62577; load the contents with file_get_contents() and parse them with simplexml_load_string().">\simplexml_load_file($dir . '/atom.xml', $cls)</error>;
    $c = <error descr="simplexml_load_file() is affected by PHP bug #62577; load the contents with file_get_contents() and parse them with simplexml_load_string().">simplexml_load_file(getenv('FEED'),  $cls,$flags, $nsUri, true)</error>;
    $d = $this->simplexml_load_file('x.xml');
    return [$a, $b, $c, $d, $empty];
}
