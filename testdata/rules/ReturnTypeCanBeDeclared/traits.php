<?php
// A trait is never a type: `new static` / `clone $this` in a trait are
// instances of the using class.
trait Singleton
{
    protected static $instance;

    public static function instance()
    {
        return static::$instance ?: static::$instance = new static();
    }

    public function copy()
    {
        return clone $this;
    }

    public function maybe($f)
    {
        if ($f) {
            return new static();
        }
        return null;
    }

    public function maybeCopy($f)
    {
        if ($f) {
            return clone $this;
        }
        return null;
    }
}
