<?php
// Called with no argument, these parent constructors do nothing: each
// statement only runs for a parameter given a value.
abstract class Generator
{
    public function __construct($options = [])
    {
        if ($options) {
            $this->setOptions($options);
        }
    }

    public function setOptions(array $options) {}
}

class DocBlock extends Generator
{
    public function __construct($short = null, $long = null, $tags = [], $flag = false, $n = 0)
    {
        if ($short !== null) {
            $this->short = $short;
        }
        if (null !== $long) {
            $this->long = $long;
        }
        if (!empty($tags)) {
            $this->tags = $tags;
        }
        if (isset($short)) {
            $this->flag = $flag;
        }
        if ($flag) {
            $this->flag = true;
        }
        if (($n)) {
            $this->n = $n;
        }
    }
}

final class FileGen extends Generator
{
    public function __construct(array $options = [])
    {
        $this->setOptions($options);
    }
}

final class Method extends DocBlock
{
    public function __construct()
    {
    }
}

// These parent constructors still do something.
abstract class Counted
{
    public function __construct($start = 1)
    {
        if ($start) {
            $this->count = $start;
        }
    }
}

final class Counter extends Counted
{
    public function <error descr="__construct does not call Counted::__construct().">__construct</error>()
    {
    }
}

abstract class Required
{
    public function __construct($name)
    {
        if ($name) {
            $this->name = $name;
        }
    }
}

final class Named extends Required
{
    public function <error descr="__construct does not call Required::__construct().">__construct</error>()
    {
    }
}

abstract class Branchy
{
    public function __construct($a = null, $b = null, $c = null, $d = null)
    {
        if ($a) {
            $this->a = $a;
        } else {
            $this->a = 1;
        }
    }
}

final class Branched extends Branchy
{
    public function <error descr="__construct does not call Branchy::__construct().">__construct</error>()
    {
    }
}

abstract class Guarded
{
    public function __construct($a = 'x', $b = null, $c = null)
    {
        if ($b === null) {
            $this->b = 1;
        }
    }
}

final class Guards extends Guarded
{
    public function <error descr="__construct does not call Guarded::__construct().">__construct</error>()
    {
    }
}
