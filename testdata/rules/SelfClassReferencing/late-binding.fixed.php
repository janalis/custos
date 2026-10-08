<?php
// self::make() forwards the late static binding; Model::make() does not.
class Model
{
    public static function make(): static { return new static(); }

    public static function build(): self
    {
        return Model::make();
    }
}

class User extends Model {}

final class Leaf
{
    public static function make(): static { return new static(); }

    public static function build(): self
    {
        return self::make();
    }
}
