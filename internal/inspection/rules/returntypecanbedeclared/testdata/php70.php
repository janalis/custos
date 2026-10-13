<?php
final class Legacy {
    // void and nullable types need PHP 7.1
    public function nothing() { }
    /** @return int|null */
    public function maybe() { return null; }
    public function <weak_warning descr="Declare ': int' as the return type.">size</weak_warning>() { return 1; }
}
