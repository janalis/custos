<?php
class Label {
    /** @param string $text */
    public function render($text) { return $text; }
}
class BoldLabel extends Label {
    public function render($text) { return $text; }
}
final class Plain {
    public function untyped($v) { return $v; }
    public function noop(): void {}
    private function me(): \Plain { return new Plain(); }
    public function self() { return $this; }
}
function plainFunction() { return 1; }
