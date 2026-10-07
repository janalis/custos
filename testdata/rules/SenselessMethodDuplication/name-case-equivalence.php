<?php
class Format { public static function text($s) { return $s; } }
class Base {
    protected $name = '';
    public function label() {
        $text = Format::Text($this->name);
        return $text;
    }
}
class Child extends Base {
    public function <weak_warning descr="Method 'label' duplicates the inherited implementation; remove it.">label</weak_warning>() {
        $text = format::text($this->name);
        return $text;
    }
}
