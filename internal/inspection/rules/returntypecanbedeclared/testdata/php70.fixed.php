<?php
final class Legacy {
    // void and nullable types need PHP 7.1
    public function nothing() { }
    /** @return int|null */
    public function maybe() { return null; }
    public function size(): int { return 1; }
}
