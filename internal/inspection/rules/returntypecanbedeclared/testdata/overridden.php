<?php
class Shape {
    /** @param int $n */
    public function <weak_warning descr="Declare ': int' as the return type (update the whole hierarchy with a signature refactoring).">sides</weak_warning>($n) { return $n; }
}
class Square extends Shape {
    public function <weak_warning descr="Declare ': int' as the return type (update the whole hierarchy with a signature refactoring).">sides</weak_warning>($n) { return 4; }
}
