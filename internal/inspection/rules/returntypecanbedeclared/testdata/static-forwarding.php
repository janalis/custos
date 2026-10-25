<?php
abstract class WidgetBase
{
    /** @return static */
    public static function <weak_warning descr="Declare ': static' as the return type (update the whole hierarchy with a signature refactoring).">create</weak_warning>(array $config)
    {
        return new static();
    }

    /** @return static */
    public static function <weak_warning descr="Declare ': static' as the return type.">fresh</weak_warning>()
    {
        return new static();
    }

    /** @return $this */
    public function <weak_warning descr="Declare ': self' as the return type (update the whole hierarchy with a signature refactoring).">copy</weak_warning>()
    {
        return $this;
    }
}

class Upload extends WidgetBase
{
    // parent::create() keeps the late static binding: static, not WidgetBase.
    public static function <weak_warning descr="Declare ': static' as the return type (update the whole hierarchy with a signature refactoring).">create</weak_warning>(array $config)
    {
        $instance = parent::create($config);
        $instance->init();
        return $instance;
    }

    public function <weak_warning descr="Declare ': static' as the return type (update the whole hierarchy with a signature refactoring).">copy</weak_warning>()
    {
        return parent::copy();
    }

    public function <weak_warning descr="Declare ': void' as the return type.">init</weak_warning>()
    {
    }
}

final class Plain extends Upload
{
    public static function <weak_warning descr="Declare ': \Upload' as the return type.">build</weak_warning>()
    {
        $made = Upload::create([]);
        return $made;
    }

    public static function <weak_warning descr="Declare ': static' as the return type.">make</weak_warning>(bool $fresh)
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

    // Nor does a method the class does not declare.
    public static function missing()
    {
        return self::nowhere();
    }

    // A class or method named by an expression says nothing.
    public static function dynamic($cls, string $m)
    {
        $made = $cls::create([]);
        $made = static::$m();
        return $made;
    }
}
