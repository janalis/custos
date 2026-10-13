<?php

#[Attribute(Attribute::TARGET_ALL)]
class Tag
{
    public function __construct(public mixed $value = null) {}
}

class Worker
{
    #[Tag(static function () { return 1; })]
    public function run() { return 'run'; }

    #[Tag(static function () { return 2; })] #[Tag]
    public function &ref() { static $v = 'ref'; return $v; }

    #[Tag(static function () { return 3; })]
    public static function make() { return 'make'; }

    #[Tag(static function () { return 4; })]
    public const int|string LIMIT = 5;

    #[Tag(static function () { return 5; })]
    public $state = 'state';
}

enum Level { case Low; }

$w = new Worker;
echo $w->run(), $w->ref(), Worker::make(), Worker::LIMIT, $w->state, "\n";
