<?php
class Report
{
    const ROW = '%s: %d%%';
    public static $line = '%-10s|%5.1f';
    private $head = '%%title%% %s';

    public function render($fh, $name, $total, $raw)
    {
        $fmt = '%%%s=%x';
        echo <error descr="This call needs 3 argument(s) in total.">sprintf</error>($fmt, $name);
        <error descr="This call needs 4 argument(s) in total.">fprintf</error>($fh, self::ROW, $name, $total, $raw);
        echo <error descr="This call needs 3 argument(s) in total.">sprintf</error>(self::$line, $name);
        echo <error descr="This call needs 2 argument(s) in total.">sprintf</error>($this->head);
        echo sprintf(<error descr="Malformed format string.">'50 %, total %s'</error>, $name);
        <error descr="This call needs 4 argument(s) in total.">sscanf</error>($raw, '%d-%d', $a);
        <error descr="This call needs 2 argument(s) in total.">printf</error>('%s');
    }
}

function scan_bare($raw)
{
    if (<error descr="This call needs 3 argument(s) in total.">sscanf</error>($raw, '%d')) {
        return 1;
    }
    return 0;
}
