<?php
class Report {
    public function build(array $cols, $fn) {
        $e = call_user_func_array([$this, 'render'], $cols);
        $f = call_user_func_array("{$this->kind}_fmt", $cols);
        $g = call_user_func_array('sprintf', $cols['main']);
        $h = call_user_func_array('sprintf');
        $i = call_user_func_array($fn, $cols);
        $j = call_user_func_array('sprintf', ($cols));
        $k = call_user_func_array('sprintf', $cols, 1);
        $l = $this->call_user_func_array('sprintf', $cols);
        // strings that cannot be written as a direct call
        $m = call_user_func_array('Report::render', $cols);
        $n = call_user_func_array('', $cols);
        $o = call_user_func_array('my fmt', $cols);
        $p = call_user_func_array('1fmt', $cols);
        $q = call_user_func_array('Lib\\', $cols);
        $r = call_user_func_array(<<<'TXT'
sprintf
TXT, $cols);
    }
}
