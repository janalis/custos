<?php
// `$this->m()` resolves m on the runtime class: the fix keeps that with
// static:: unless m cannot be overridden.
class Report {
    public static function title() { return 'report'; }
    private static function slug() { return 'r'; }
    final public static function stamp() { return 1; }

    public function header() {
        return [
            <warning descr="Static method title() called through $this; use static::title().">$this</warning>->title(),
            <warning descr="Static method slug() called through $this; use self::slug().">$this</warning>->slug(),
            <warning descr="Static method stamp() called through $this; use self::stamp().">$this</warning>->stamp(),
        ];
    }
}

final class Invoice {
    public static function number() { return 7; }
    public function label() { return <warning descr="Static method number() called through $this; use self::number().">$this</warning>->number(); }
}

enum Unit {
    case Metre;
    public static function base() { return self::Metre; }
    public function root() { return <warning descr="Static method base() called through $this; use self::base().">$this</warning>->base(); }
}

// An abstract method of a trait only states what the using class must
// provide; the implementation inherited from a parent decides. Inside the
// trait the requirement is all there is (it must be static).
trait NeedsAsserts {
    abstract public static function assertSame($a, $b);
    public function check() { <warning descr="Static method assertSame() called through $this; use static::assertSame().">$this</warning>->assertSame(1, 1); }
}

// The parent is not indexed: its assertSame() is unknown.
final class RemoteCase extends \Vendor\Missing\TestCase {
    use NeedsAsserts;
    public function run() { $this->assertSame(2, 2); }
}

class LocalBase {
    public static function assertSame($a, $b) { return $a === $b; }
}

// The inherited implementation is found and reported.
final class LocalCase extends LocalBase {
    use NeedsAsserts;
    public function run() { return <warning descr="Static method assertSame() called through $this; use self::assertSame().">$this</warning>->assertSame(3, 3); }
}

abstract class Shape {
    abstract public static function sides();
    public function describe() { return <warning descr="Static method sides() called through $this; use static::sides().">$this</warning>->sides(); }
}
