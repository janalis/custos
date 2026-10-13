<?php
namespace App;

/** @method static Query where(string $col, mixed $val) */
class User
{
    public function __call($name, $args) { return (new Query())->$name(...$args); }
    public static function __callStatic($name, $args) { return (new static())->$name(...$args); }

    public function active(): Query
    {
        return $this->where('active', 1);
    }
}

class Query
{
    public function where($col, $val) { return $this; }
}

function scoped(User $user): void
{
    $copy = $user;
    $copy->where('a', 1);
}
