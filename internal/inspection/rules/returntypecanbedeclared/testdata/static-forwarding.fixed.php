<?php
abstract class WidgetBase
{
    /** @return static */
    public static function create(array $config)
    {
        return new static();
    }

    /** @return static */
    public static function fresh(): static
    {
        return new static();
    }

    /** @return $this */
    public function copy()
    {
        return $this;
    }
}

class Upload extends WidgetBase
{
    // parent::create() keeps the late static binding: static, not WidgetBase.
    public static function create(array $config)
    {
        $instance = parent::create($config);
        $instance->init();
        return $instance;
    }

    public function copy()
    {
        return parent::copy();
    }

    public function init(): void
    {
    }
}

final class Plain extends Upload
{
    public static function build()
    {
        $made = Upload::create([]);
        return $made;
    }

    public static function make(bool $fresh)
    {
        $made = static::fresh();
        if ($fresh) {
            $made = self::fresh();
        }
        return $made;
    }

    public static function mixedUp(bool $fresh)
    {
        $made = static::fresh();
        if ($fresh) {
            $made = $fresh;
        }
        return $made;
    }

    // A class or method named by an expression says nothing.
    public static function dynamic($cls, string $m)
    {
        $made = $cls::create([]);
        $made = static::$m();
        return $made;
    }
}
