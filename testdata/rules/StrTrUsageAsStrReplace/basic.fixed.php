<?php
class Slugger {
    const SEP = '.';
    private $glue = ':';
    public function make($title, $glue = '+', $flag = false) {
        $a = str_replace('-', '_', $title);
        $b = \str_replace("\t", ';', $title);
        $c = str_replace('\'', '`', $title);
        $d = str_replace("\$", 'S', $title);
        $e = str_replace($glue, ' ', $title);
        $f = str_replace(self::SEP, ' ', $title);
        $g = str_replace($this->glue, ' ', $title);
        $h = str_replace($flag ? '!' : null, ' ', $title);
        $local = '#';
        $i = str_replace($local, ' ', $title);
        return [$a, $b, $c, $d, $e, $f, $g, $h, $i];
    }
}
