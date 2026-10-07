<?php
class Printer
{
    public static $tpl;

    public function render(bool $more, $a)
    {
        $format = '50 %, total';
        if ($more) {
            $format .= ' %s';
        }
        echo sprintf($format, $a);
        echo sprintf($more ? $format : 'none', $a);
        printf(self::$tpl);
    }

    public function __construct()
    {
        $t = 'x=%s';
        $t .= ' y=%s';
        self::$tpl = $t;
    }
}
