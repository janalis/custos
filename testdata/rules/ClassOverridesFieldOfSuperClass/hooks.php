<?php
abstract class Shape {
    abstract public string $label { get; }
    public string $name = 'shape';
    public int $sides = 0;
    public string $area = '' { get => $this->area . 'm2'; }
    public $open;
    public $sealed;
    public $guarded;
    public protected(set) int $count = 0;
    public private(set) int $serial = 0;
    protected $tagged;
}

class Square extends Shape {
    // Implements the abstract property: not droppable.
    public string $label;
    // Adds hooks.
    public string $name = 'shape' { get => strtoupper($this->name); }
    public int $sides = 0 { set { $this->sides = $value * 2; } }
    // Replaces the inherited hooks with a plain backed property.
    public string $area = '';
    // Adds finality or narrows the write visibility.
    final public $open;
    public private(set) $sealed;
    public protected(set) $guarded;
    // Widens the write visibility back.
    public public(set) int $count = 0;
    // Carries an attribute the inherited declaration does not have.
    #[Assert\NotBlank]
    protected $tagged;
}

class Circle extends Shape {
    public public(set) <weak_warning descr="Property 'open' is already declared in \Shape; drop this re-declaration.">$open</weak_warning>;
    public protected(set) int <weak_warning descr="Property 'count' is already declared in \Shape; drop this re-declaration.">$count</weak_warning> = 0;
    public private(set) int <weak_warning descr="Property 'serial' is already declared in \Shape; drop this re-declaration.">$serial</weak_warning> = 0;
}
