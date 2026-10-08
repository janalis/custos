<?php
interface Shape {}
interface Named {}
class Base implements Named {}
class Circle extends Base implements Shape, Named {}
class Label {}

abstract class Fluent
{
    abstract public function skip(int $n);

    public function none() { return 1; }

    public function with(): static { return $this; }

    public static function make(): static { return new static(); }

    public function self(self $me, Shape $s, Label $l, Ghost $g)
    {
        $me = <warning descr="Assigning a value of type \Label does not match the parameter's declared type.">new Label()</warning>;
        $s = <warning descr="Assigning a value of type \Fluent does not match the parameter's declared type.">$this->with()</warning>;
        $s = <warning descr="Assigning a value of type \Fluent does not match the parameter's declared type.">Fluent::make()</warning>;
        $s = <warning descr="Assigning a value of type \Fluent does not match the parameter's declared type.">self::make()</warning>;
        $s = Missing::make();
        $s = <warning descr="Assigning a value of type \Fluent does not match the parameter's declared type.">$s ?? $this->with()</warning>;
        $l = $l->with();
        $g = new Label();
        $s = new circle();
        return [is_object($s), is_a($s, Shape::class)];
    }
}

/**
 * @param static $st
 */
function closures(\Closure $c, $st, Shape $shape, int $n, string $str)
{
    $ok = is_callable($c) && is_string($c) && is_array($c);
    $shape = rand() ? new Circle() : null;
    switch ($n) {
        case 1:
            $n = <warning descr="Assigning a value of type string does not match the parameter's declared type.">'one'</warning>;
            break;
        default:
            $n = <warning descr="Assigning a value of type string does not match the parameter's declared type.">'two'</warning>;
    }
    if ($n) {
        $str = 'a';
        {
            return $ok;
        }
        $str = 1;
    }
    if ($n > 1) {
        throw new Exception();
    } elseif ($n > 2) {
        return 1;
    } else {
        exit();
    }
    $str = 2;
}

function partial(string $str, int $n)
{
    if ($n > 1) {
        return 1;
    } elseif ($n > 2) {
        $str = <warning descr="Assigning a value of type int does not match the parameter's declared type.">5</warning>;
    } else {
        return 3;
    }
    throw new Exception(<warning descr="This check is always false for the declared parameter type; is the parameter being reused?">is_int($str)</warning>);
    $str = 2;
}

function noparams() { return is_int(1); }

function broken(int , string $s) { return <warning descr="This check is always false for the declared parameter type; is the parameter being reused?">is_int($s)</warning>; }

function fail(): never { exit(1); }

function nev(int $n, self $outside) {
    $n = fail();
    $outside = <warning descr="Assigning a value of type \Label does not match the parameter's declared type.">new Label()</warning>;
}

class Child extends Fluent
{
    /**
     * @param static $st
     */
    public function child(Shape $s, $st, $obj, Circle $c)
    {
        $s = <warning descr="Assigning a value of type \Fluent does not match the parameter's declared type.">parent::make()</warning>;
        $s = <warning descr="Assigning a value of type \Child does not match the parameter's declared type.">new static()</warning>;
        $st = $this->with();
        $st = $st ?? Missing::make();
        $st = $st ?? $obj::make();
        $st = $st ?? $obj->make();
        $st = $other ?? $obj->make();
        $f = new Child();
        $s = <warning descr="Assigning a value of type \Child does not match the parameter's declared type.">$f::make()</warning>;
        $w = rand() ? new Child() : (rand() ? new Kid2() : new Kid3());
        $s = $w->with();
        $s = $this->with() ?? new Circle();
        $c = new circle();
    }
}

class Orphan
{
    /**
     * @param static $st
     */
    public function me(): static { return $this; }

    public function orphan($st)
    {
        $st = $this->me();
        $st = $st ?? parent::make();
    }
}

function anon() {
    return new class extends Fluent {
        public function in(Shape $s) {
            $s = self::make();
            $s = static::make();
        }
    };
}

final class Kid2 extends Fluent {}
final class Kid3 extends Fluent {}
