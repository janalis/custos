<?php
class Slugger {
    const SEP = '.';
    private $glue = ':';
    public function make($title, $glue = '+', $flag = false) {
        $a = <weak_warning descr="Use 'str_replace('-', '_', $title)' instead.">strtr($title, '-', '_')</weak_warning>;
        $b = <weak_warning descr="Use '\str_replace(&quot;\t&quot;, ';', $title)' instead.">\strtr($title, "\t", ';')</weak_warning>;
        $c = <weak_warning descr="Use 'str_replace('\'', '`', $title)' instead.">strtr($title, '\'', '`')</weak_warning>;
        $d = <weak_warning descr="Use 'str_replace(&quot;\$&quot;, 'S', $title)' instead.">strtr($title, "\$", 'S')</weak_warning>;
        $e = <weak_warning descr="Use 'str_replace($glue, ' ', $title)' instead.">strtr($title, $glue, ' ')</weak_warning>;
        $f = <weak_warning descr="Use 'str_replace(self::SEP, ' ', $title)' instead.">strtr($title, self::SEP, ' ')</weak_warning>;
        $g = <weak_warning descr="Use 'str_replace($this->glue, ' ', $title)' instead.">strtr($title, $this->glue, ' ')</weak_warning>;
        $h = <weak_warning descr="Use 'str_replace($flag ? '!' : null, ' ', $title)' instead.">strtr($title, $flag ? '!' : null, ' ')</weak_warning>;
        $local = '#';
        $i = <weak_warning descr="Use 'str_replace($local, ' ', $title)' instead.">strtr($title, $local, ' ')</weak_warning>;
        return [$a, $b, $c, $d, $e, $f, $g, $h, $i];
    }
}
