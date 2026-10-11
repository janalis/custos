<?php

namespace Depot;

class Crate {}
class Pallet extends Crate {}

class Node
{
    public function root(): self { return new Node(); }
    public function same(): static { return $this; }
}
class Leaf extends Node {}

interface Linked
{
    public function next(): self;
}
final class Chain implements Linked
{
    public function next(): self { return $this; }
}

trait Fluent
{
    public function tap(): self { return $this; }
    public static function make(): self { return new static(); }

    public function check()
    {
        $this->assertInstanceOf(Fluent::class, $this->tap());
        $this->assertInternalType('object', $this->tap());
    }
}
class Builder { use Fluent; }
class SubBuilder extends Builder {}

class Request
{
    public function method(): string { return 'GET'; }
    public function path(): string { return '/'; }
}
/** @method string method() */
class Dynamic
{
    public function __call($name, $args) { return 'x'; }
}
class Sink
{
    public function __construct(&$slot) { $slot = 1; }
}
class Holder
{
    public function __construct($value) {}
}
class Bare {}

function crate(): Crate { return new Crate(); }
function nothing(): void {}
function fill(&$slot): void { $slot = 'raw'; }
function peek($value): void {}

abstract class DepotTest
{
    private $kept;

    abstract protected function flush(): void;

    public function testSelf(Leaf $leaf, Chain $chain, Builder $builder, SubBuilder $sub)
    {
        // `self` is the declaring class, `static` the receiver's.
        $this->assertInstanceOf(Leaf::class, $leaf->root());
        $this->assertInstanceOf(Node::class, $leaf->root());
        $this->assertInstanceOf(Leaf::class, $leaf->same());
        $this->assertInstanceOf(Chain::class, $chain->next());
        // In a trait, `self` is the class importing it.
        $this->assertInstanceOf(Builder::class, $builder->tap());
        $this->assertInstanceOf(Builder::class, $sub->tap());
        $this->assertInstanceOf(SubBuilder::class, $sub->tap());
        $this->assertInstanceOf(Fluent::class, $builder->tap());
        // Reached without an importing class, `self` stays unresolved.
        $this->assertInstanceOf(Fluent::class, Fluent::make());
    }

    public function testCallable(Leaf $leaf)
    {
        $this->assertNull(nothing(...));
        $this->assertNull($this->flush(...));
        $this->assertInstanceOf(Leaf::class, $leaf->same(...));
        $this->assertInstanceOf(Crate::class, crate(...));
    }

    public function testRebound(array $rows, $unknown)
    {
        $a = crate();
        foreach ($rows as $a) {
            $this->assertInstanceOf(Crate::class, $a);
        }
        $b = crate();
        fill($b);
        $this->assertInstanceOf(Crate::class, $b);
        $c = nothing();
        [$c] = $rows;
        $this->assertNull($c);
        $d = nothing();
        try {
            nothing();
        } catch (\Exception $d) {
            $this->assertNull($d);
        }
        $e = nothing();
        $fn = function () use (&$e) { $e = 1; };
        $this->assertNull($e);
        global $f;
        $f = nothing();
        $this->assertNull($f);
        static $g;
        $g = nothing();
        $this->assertNull($g);
        $h = nothing();
        unset($h);
        $this->assertNull($h);
        $i = nothing();
        $i[] = 1;
        $this->assertNull($i);
        $j = nothing();
        $alias = &$j;
        $this->assertNull($j);
        $k = nothing();
        $refs = [&$k];
        $this->assertNull($k);
        $l = crate();
        $unknown->load($l);
        $this->assertInstanceOf(Crate::class, $l);
        $m = crate();
        new Sink($m);
        $this->assertInstanceOf(Crate::class, $m);
        $n = crate();
        new Missing($n);
        $this->assertInstanceOf(Crate::class, $n);
        $o = crate();
        new $unknown($o);
        $this->assertInstanceOf(Crate::class, $o);
        $p = crate();
        new class($p) {};
        $this->assertInstanceOf(Crate::class, $p);
        $q = $r = crate();
        fill($r);
        $s = $q;
        $t = $r;
        $this->assertInstanceOf(Crate::class, $s);
        $this->assertInstanceOf(Crate::class, $t);
        $u = nothing();
        $u .= 'x';
        $this->kept = $u;
        $this->assertNull($this->kept);
    }

    public function testKept(array $rows)
    {
        // By-value uses, element reads and property writes do not rebind.
        $a = crate();
        peek($a);
        $a->label = 'x';
        $a->tags['k'] = 1;
        $copy = function () use ($a) { return $a; };
        $list = [$a];
        $rows[$a] = 1;
        new Holder($a);
        new Bare($a);
        $this->assertNotNull($a);
        $this->assertInstanceOf(Crate::class, $a);
        $b = nothing();
        $rows[$b] = 1;
        $other = &$rows;
        $this->assertNull($b);
    }

    public function testMatchers(?Request $maybe)
    {
        // No fix: the doubled class declares its own method().
        $request = $this->createMock(Request::class);
        $request->expects($this->any())->method('path')->willReturn('/home');
        $built = $this->getMockBuilder(Request::class)->disableOriginalConstructor()->getMock();
        $built->expects($this->any())->method('path');
        // No fix: class reference or nullsafe receiver.
        static::expects(self::any())->method('x');
        $maybe?->expects($this->any())->method('x');
        // Fixed: `method` is only a documented magic method there.
        $dynamic = $this->createMock(Dynamic::class);
        $dynamic->method('x');
        $crate = $this->createMock(Crate::class);
        $crate->method('x');
    }
}
