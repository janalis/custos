<?php
class Registry {
    public static $map = [];
    public $slot;
    /** @param Registry[] $list */
    public function probe($name, $cls, array $list) {
        $a = isset($this->$name);                  // dynamic name
        $b = isset(Registry::$$name);              // variable static name
        $c = Registry::$map !== null;
        $d = isset($cls::$map);                    // unknown class
        $e = isset($list['k']);
        return [$a, $b, $c, $d, $e];
    }
}
