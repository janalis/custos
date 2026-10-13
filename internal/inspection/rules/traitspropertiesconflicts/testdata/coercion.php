<?php
trait Measures {
    public float $ratio = 1.0;
    public ?float $scale = 2.0;
    public float|string $step = 3.0;
    public int|float $count = 4.0;
    public $loose = 5.0;
    public $zero = 0.0;
    public float $origin = 0.0;
    public float $offset = -6.0;
    public float $limit = 7.0;
    public iterable $items = [];
    public ?iterable $rows = null;
    public \Traversable|array $cells = [];
    public iterable $names = [];
}

class Gauge {
    use Measures;
    public float <weak_warning descr="Gauge and trait Measures both declare property $ratio.">$ratio</weak_warning> = 1;
    public ?float <weak_warning descr="Gauge and trait Measures both declare property $scale.">$scale</weak_warning> = 2;
    public float|string <weak_warning descr="Gauge and trait Measures both declare property $step.">$step</weak_warning> = 3;
    public int|float <error descr="Gauge and trait Measures both declare property $count.">$count</error> = 4;
    public <error descr="Gauge and trait Measures both declare property $loose.">$loose</error> = 5;
    public <weak_warning descr="Gauge and trait Measures both declare property $zero.">$zero</weak_warning> = -0.0;
    public float <weak_warning descr="Gauge and trait Measures both declare property $origin.">$origin</weak_warning> = -0;
    public float <weak_warning descr="Gauge and trait Measures both declare property $offset.">$offset</weak_warning> = -6;
    public float <error descr="Gauge and trait Measures both declare property $limit.">$limit</error> = 8;
    public \Traversable|array <weak_warning descr="Gauge and trait Measures both declare property $items.">$items</weak_warning> = [];
    public array|\Traversable|null <weak_warning descr="Gauge and trait Measures both declare property $rows.">$rows</weak_warning> = null;
    public iterable <weak_warning descr="Gauge and trait Measures both declare property $cells.">$cells</weak_warning> = [];
    public array <error descr="Gauge and trait Measures both declare property $names.">$names</error> = [];
}
