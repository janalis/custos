<?php
// self::make() forwards the late static binding; Model::make() does not.
class Model
{
    public static function make(): static { return new static(); }

    public static function build(): <weak_warning descr="Refer to the class as 'self' instead of 'Model'.">Model</weak_warning>
    {
        return <weak_warning descr="Refer to the class as 'self' instead of 'Model'.">Model</weak_warning>::make();
    }
}

class User extends Model {}

final class Leaf
{
    public static function make(): static { return new static(); }

    public static function build(): <weak_warning descr="Refer to the class as 'self' instead of 'Leaf'.">Leaf</weak_warning>
    {
        return <weak_warning descr="Refer to the class as 'self' instead of 'Leaf'.">Leaf</weak_warning>::make();
    }
}
