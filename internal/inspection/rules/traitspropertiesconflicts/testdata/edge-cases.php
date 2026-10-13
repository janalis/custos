<?php
namespace Shop;

interface Pricing {}

trait Price {
    public $amount = -1;
    public $big = 9223372036854775808;
    public $rate = -0.5;
    public $huge = 1e999;
    public $esc = 'a\'b';
    public $neg = -FOO;
    public $tag = -'x';
    public ?Pricing $pricing = null;
    public ?self $self = null;
    public $sum = 1 + 1;
    public $only = 'kept';
    public $hex = -0x10;
}

trait Label {
    public $title = 'x';
}

class NotATrait {}

class Product {
    use Price, Label {
        Price::amount insteadof Label;
        Label::title as heading;
    }
    use NotATrait;
    public <weak_warning descr="Product and trait Price both declare property $amount.">$amount</weak_warning> = -1;
    public <error descr="Product and trait Price both declare property $big.">$big</error> = 9223372036854775807;
    public <weak_warning descr="Product and trait Price both declare property $rate.">$rate</weak_warning> = -0.50;
    public $huge = 2e999;
    public $esc = "a'b";
    public $neg = -BAR;
    public $tag = -'y';
    public ?Pricing <weak_warning descr="Product and trait Price both declare property $pricing.">$pricing</weak_warning> = null;
    public ?self $self = NULL;
    public $sum = 2;
    public <weak_warning descr="Product and trait Price both declare property $hex.">$hex</weak_warning> = -16;

    /** @ORM\Column */
    public $only = 'other';

    public function __construct($plain) {}

    public function other() {}
}

class Orphan extends MissingBase {
    use Price;
}

class Holder {
    public $unrelated = 1;
}

class Leaf extends Holder {
    use Price;
}

trait Loop1 { use Loop2; }
trait Loop2 { use Loop1; }

class Looped {
    use Loop1;
    public $none = 1;
}
