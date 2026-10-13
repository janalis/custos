<?php
class Report {
    public function build(array $cols) {
        $a = call_user_func_array('sprintf', $cols);
        $b = \call_user_func_array("max", $this->totals);
        $c = call_user_func_array('min', [3, 9]);
        $d = call_user_func_array('\\Lib\\fmt', self::defaults());
        $e = call_user_func_array('implode', array($sep, $cols));
        $f = call_user_func_array('trim', self::$cache);
        $g = call_user_func_array('abs', $this->load());
    }
}
