<?php
// Model::make() resets the late static binding that self::make() forwards.
class Model
{
    public static function make() { return new static; }
    public static function name() { return static::class; }

    public static function build()
    {
        $limit = Model::LIMIT;
        return self::make();
    }

    public function label() { return self::name(); }
}

class User extends Model {}

final class Leaf
{
    public static function make() { return new static; }

    public static function build()
    {
        return Leaf::make();
    }
}
