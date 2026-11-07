<?php
function loadedValue($value) {
    <error descr="Parameter is overwritten before its value is used.">$value</error> = load_value();
    return $value;
}
class Loader {
    public function reload($value) {
        <error descr="Parameter is overwritten before its value is used.">$value</error> = $this->fetch();
        return $value;
    }
    private function fetch() {
        return random_int(0, 9);
    }
}
function introspectedAfterLoad($value) {
    <error descr="Parameter is overwritten before its value is used.">$value</error> = load_value();
    return [$value, get_defined_vars()];
}
function loopedLoad($value) {
    while (more_values()) {
        $value = load_value();
    }
    return $value;
}
function skippedLoad($value, $condition) {
    if ($condition) {
        $value = load_value();
    }
    return $value;
}
