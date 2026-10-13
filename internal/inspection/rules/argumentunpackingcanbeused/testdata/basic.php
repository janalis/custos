<?php
class Report {
    public function build(array $cols) {
        $a = <warning descr="Call 'sprintf(...$cols)' directly using argument unpacking (wrap with array_values() when keys are not sequential).">call_user_func_array('sprintf', $cols)</warning>;
        $b = <warning descr="Call 'max(...$this->totals)' directly using argument unpacking (wrap with array_values() when keys are not sequential).">\call_user_func_array("max", $this->totals)</warning>;
        $c = <warning descr="Call 'min(...[3, 9])' directly using argument unpacking (wrap with array_values() when keys are not sequential).">call_user_func_array('min', [3, 9])</warning>;
        $d = <warning descr="Call '\Lib\fmt(...self::defaults())' directly using argument unpacking (wrap with array_values() when keys are not sequential).">call_user_func_array('\\Lib\\fmt', self::defaults())</warning>;
        $e = <warning descr="Call 'implode(...array($sep, $cols))' directly using argument unpacking (wrap with array_values() when keys are not sequential).">call_user_func_array('implode', array($sep, $cols))</warning>;
        $f = <warning descr="Call 'trim(...self::$cache)' directly using argument unpacking (wrap with array_values() when keys are not sequential).">call_user_func_array('trim', self::$cache)</warning>;
        $g = <warning descr="Call 'abs(...$this->load())' directly using argument unpacking (wrap with array_values() when keys are not sequential).">call_user_func_array('abs', $this->load())</warning>;
    }
}
