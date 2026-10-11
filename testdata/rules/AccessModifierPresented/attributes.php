<?php

#[Attribute(Attribute::TARGET_ALL)]
class Tag
{
    public function __construct(public mixed $value = null) {}
}

class Worker
{
    #[Tag(static function () { return 1; })]
    function <weak_warning descr="Declare the visibility of 'run' explicitly.">run</weak_warning>() { return 'run'; }

    #[Tag(static function () { return 2; })] #[Tag]
    function &<weak_warning descr="Declare the visibility of 'ref' explicitly.">ref</weak_warning>() { static $v = 'ref'; return $v; }

    #[Tag(static function () { return 3; })]
    static function <weak_warning descr="Declare the visibility of 'make' explicitly.">make</weak_warning>() { return 'make'; }

    #[Tag(static function () { return 4; })]
    const int|string <weak_warning descr="Declare the visibility of 'LIMIT' explicitly.">LIMIT</weak_warning> = 5;

    #[Tag(static function () { return 5; })]
    var <weak_warning descr="Declare the visibility of 'state' explicitly.">$state</weak_warning> = 'state';
}

enum Level { case Low; }

$w = new Worker;
echo $w->run(), $w->ref(), Worker::make(), Worker::LIMIT, $w->state, "\n";
