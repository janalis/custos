<?php
class Registry {
    public static $map = [];
    public $slot;
    /** @param Registry[] $list */
    public function probe($name, $cls, array $list) {
        $a = isset($this->$name);                  // dynamic name
        $b = isset(Registry::$$name);              // variable static name
        $c = <weak_warning descr="Compare with null instead: 'Registry::$map !== null'.">isset(Registry::$map)</weak_warning>;
        $d = isset($cls::$map);                    // unknown class
        $e = isset(<weak_warning descr="Use array_key_exists() to check for the key itself.">$list['k']</weak_warning>);
        return [$a, $b, $c, $d, $e];
    }
}
