<?php
class Separators {
    const DOT = '.';
    const ALIAS = self::DOT;
    const WRAPPED = ('.');
    public $ch = ',';
    public function run($s, array $cfg, array $list, $glue = '+') {
        $a = str_replace(self::DOT, '_', $s);
        $b = str_replace('-', SELF::DOT, $s);
        $c = str_replace(('-'), ('_'), $s);
        $d = strtr($s, static::DOT, '_');
        $e = strtr($s, self::ALIAS, '_');
        $f = function ($s) {
            return strtr($s, self::DOT, '_');
        };
        $g = strtr($s, Separators::DOT, '_');
        $from = $cfg['from'];
        if (!$from) {
            $from = '-';
        }
        $h = strtr($s, $from, '_');
        $i = strtr($s, $this->ch, ';');
        $each = '-';
        foreach ($list as $each) {
            $j = strtr($s, $each, '_');
        }
        $k = strtr($s, $glue, ' ');
        return [$a, $b, $c, $d, $e, $f, $g, $h, $i, $j, $k];
    }
}
