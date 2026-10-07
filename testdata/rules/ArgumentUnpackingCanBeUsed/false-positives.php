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
    }
}
