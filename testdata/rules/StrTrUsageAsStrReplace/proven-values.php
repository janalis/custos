<?php
class Separators {
    const DOT = '.';
    const ALIAS = self::DOT;
    const WRAPPED = ('.');
    public $ch = ',';
    public function run($s, array $cfg, array $list, $glue = '+') {
        $a = <weak_warning descr="Use 'str_replace(self::DOT, '_', $s)' instead.">strtr($s, self::DOT, '_')</weak_warning>;
        $b = <weak_warning descr="Use 'str_replace('-', SELF::DOT, $s)' instead.">strtr($s, '-', SELF::DOT)</weak_warning>;
        $c = <weak_warning descr="Use 'str_replace(('-'), ('_'), $s)' instead.">strtr($s, ('-'), ('_'))</weak_warning>;
        $d = <weak_warning descr="Use 'str_replace(static::DOT, '_', $s)' instead.">strtr($s, static::DOT, '_')</weak_warning>;
        $e = <weak_warning descr="Use 'str_replace(self::ALIAS, '_', $s)' instead.">strtr($s, self::ALIAS, '_')</weak_warning>;
        $f = function ($s) {
            return <weak_warning descr="Use 'str_replace(self::DOT, '_', $s)' instead.">strtr($s, self::DOT, '_')</weak_warning>;
        };
        $g = <weak_warning descr="Use 'str_replace(Separators::DOT, '_', $s)' instead.">strtr($s, Separators::DOT, '_')</weak_warning>;
        $from = $cfg['from'];
        if (!$from) {
            $from = '-';
        }
        $h = <weak_warning descr="Use 'str_replace($from, '_', $s)' instead.">strtr($s, $from, '_')</weak_warning>;
        $i = <weak_warning descr="Use 'str_replace($this->ch, ';', $s)' instead.">strtr($s, $this->ch, ';')</weak_warning>;
        $each = '-';
        foreach ($list as $each) {
            $j = <weak_warning descr="Use 'str_replace($each, '_', $s)' instead.">strtr($s, $each, '_')</weak_warning>;
        }
        $k = <weak_warning descr="Use 'str_replace($glue, ' ', $s)' instead.">strtr($s, $glue, ' ')</weak_warning>;
        return [$a, $b, $c, $d, $e, $f, $g, $h, $i, $j, $k];
    }
}
