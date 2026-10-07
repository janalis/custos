<?php
class Label {
    /** @param string $text */
    public function <weak_warning descr="Declare ': string' as the return type (update the whole hierarchy with a signature refactoring).">render</weak_warning>($text) { return $text; }
}
class BoldLabel extends Label {
    public function <weak_warning descr="Declare ': string' as the return type (update the whole hierarchy with a signature refactoring).">render</weak_warning>($text) { return $text; }
}
final class Plain {
    public function untyped($v) { return $v; }
    public function <weak_warning descr="Declare ': void' as the return type.">noop</weak_warning>(){}
    private function <weak_warning descr="Declare ': \Plain' as the return type.">me</weak_warning>() { return new Plain(); }
    public function self() { return $this; }
}
function plainFunction() { return 1; }
