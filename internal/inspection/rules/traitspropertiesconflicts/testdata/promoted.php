<?php
trait Wraps {
    private array $items;
    private int $count = 0;
}

// Promotion has to declare the property: a compatible pair is not reported.
class Wrapper {
    use Wraps;

    public function __construct(private array $items) {}
}

// An incompatible one still is.
class Counter {
    use Wraps;

    public function __construct(private string <error descr="Counter and trait Wraps both declare property $count.">$count</error>) {}
}
