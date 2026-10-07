<?php
trait HasColor {
    public $color = 'red';
    protected $shade = 1;
}
trait HasSize {
    public $size = 10;
    private $unit = 'px';
}

/**
 * @property $note
 */
trait HasMeta {
    /** @Column */
    public $tag = 'x';
    public $label = 'L';
    protected $weight = 2;
}

class Shape {
    public $color = 'red';
    protected $size = 12;
    private $unit = 'em';
}

/**
 * @property $weight
 */
class Box extends Shape {
    use <weak_warning descr="Box and trait HasColor both declare property $color.">HasColor</weak_warning>, <error descr="Box and trait HasSize both declare property $size.">HasSize</error>;
    use HasMeta;

    /** @Column */
    public $tag = 'x';
    public <weak_warning descr="Box and trait HasMeta both declare property $label.">$label</weak_warning> = 'L';
    protected <error descr="Box and trait HasMeta both declare property $weight.">$weight</error> = 3;
    public $note = 'n';
}
