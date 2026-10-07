<?php

namespace PHPUnit\Framework {
    abstract class Assert
    {
        public static function assertEquals($expected, $actual, string $message = '') {}
    }
    abstract class TestCase extends Assert {}
}

namespace Shop\Compare {
    final class LenientComparator
    {
        public function assertEquals($left, $right, $tolerance = 0.0, $sorted = false, $folded = false): void {}
        public static function check($left, $right, $tolerance = 0.0, $sorted = false): void {}
    }

    final class Matcher
    {
        public static function assertEquals($left, $right, $tolerance = 0.0, $sorted = false): void {}
    }

    final class Runner extends \PHPUnit\Framework\TestCase
    {
        public function run(LenientComparator $c, $a, $b): void
        {
            $c->assertEquals($a, $b, 0.5, true, false);
            Matcher::assertEquals($a, $b, 0.5, true);
            $this->assertEquals($a, $b, '', <weak_warning descr="PHPUnit 8.0 deprecated the delta argument; call assertEqualsWithDelta() instead.">0.5</weak_warning>);
            $unknown->assertEquals($a, $b, 0.5, true, false);
            $this->factory()->assertEquals($a, $b, 0.5, true, false);
            $self = $this;
            $self->assertEquals($a, $b, '', <weak_warning descr="PHPUnit 8.0 deprecated the delta argument; call assertEqualsWithDelta() instead.">0.5</weak_warning>);
            $dyn::assertEquals($a, $b, '', <weak_warning descr="PHPUnit 8.0 deprecated the delta argument; call assertEqualsWithDelta() instead.">0.5</weak_warning>);
        }
    }
}
