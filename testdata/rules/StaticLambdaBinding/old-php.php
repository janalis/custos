<?php
class Widget {
    public function handlers() {
        return static function () { return $this; };
    }
}
