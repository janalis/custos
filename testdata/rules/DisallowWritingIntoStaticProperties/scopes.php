<?php

trait Counts
{
    public static $count = 0;
}

trait Tagged
{
    use Counts;
}

trait Labelled
{
    use Counts;
}

class Base
{
    public static $shared = 0;
}

class Child extends Base
{
    use Tagged, Labelled, MissingTrait;

    public function bump()
    {
        Child::$count = 1;
        <weak_warning descr="Modify this static property only from the class that declares it.">Child::$shared = 2</weak_warning>;
    }
}

class Orphan
{
    public function write()
    {
        parent::$shared = 3;
    }
}

$anon = new class {
    public function write()
    {
        <weak_warning descr="Modify this static property only from the class that declares it.">Base::$shared = 4</weak_warning>;
    }
};

$counted = new class {
    use Tagged;

    public function write()
    {
        Counts::$count = 5;
        static::$count = 6;
        <weak_warning descr="Modify this static property only from the class that declares it.">Base::$shared = 7</weak_warning>;
    }
};

$derived = new class extends Base {
    public function write()
    {
        <weak_warning descr="Modify this static property only from the class that declares it.">parent::$shared = 8</weak_warning>;
    }
};
