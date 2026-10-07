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
    }
