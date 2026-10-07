<?php
trait Counts {
    public static $count = 0;
    public function bump() { static::$count = 1; }
}
class Page {
    use Counts;
    public function reset() { static::$count = 0; Page::$count = 0; }
}
self::$x = 1;
