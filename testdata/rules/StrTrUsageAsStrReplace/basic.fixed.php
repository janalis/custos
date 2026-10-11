<?php
class Slugger {
    const SEP = '.';
    private $glue = ':';
    public function make($title, $glue = '+', $flag = false) {
        $a = str_replace('-', '_', $title);
        $b = \str_replace("\t", ';', $title);
        $c = str_replace('\'', '`', $title);
        $d = str_replace("\$", 'S', $title);
        $e = strtr($title, $glue, ' ');
        $f = str_replace(self::SEP, ' ', $title);
        $g = strtr($title, $this->glue, ' ');
        $h = strtr($title, $flag ? '!' : null, ' ');
        $local = '#';
        $i = strtr($title, $local, ' ');
        return [$a, $b, $c, $d, $e, $f, $g, $h, $i];
    }
}
