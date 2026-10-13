<?php
class Report {
    public function build(array $cols) {
        $a = sprintf(...$cols);
        $b = max(...$this->totals);
        $c = min(...[3, 9]);
        $d = \Lib\fmt(...self::defaults());
        $e = implode(...array($sep, $cols));
        $f = trim(...self::$cache);
        $g = abs(...$this->load());
    }
}
