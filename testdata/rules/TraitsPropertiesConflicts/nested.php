<?php
trait HasColor {
    public $color = [1,2];
}
trait Nested {
    use HasColor;
}
class Base {
    public $color = [1, 2];
}
class Child extends Base {
    use <weak_warning descr="Child and trait Nested both declare property $color.">Nested</weak_warning>;
}
class Other {
    use Nested;
    public <weak_warning descr="Other and trait Nested both declare property $color.">$color</weak_warning> = [ 1, 2 ];
}
