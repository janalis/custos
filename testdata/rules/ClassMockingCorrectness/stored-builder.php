<?php

namespace PHPUnit\Framework {
    class TestCase
    {
        public function getMockBuilder($c) {}
    }
}

namespace {
    abstract class Req { abstract public function name(): string; }
    trait Mixin {}

    class ReqTest extends \PHPUnit\Framework\TestCase
    {
        private $builder;

        public function testStored(): void
        {
            $mb = $this->getMockBuilder(Req::class);
            $req = $mb->getMockForAbstractClass();
            $tb = $this->getMockBuilder(Mixin::class);
            $tb->getMockForTrait();
        }

        public function testEscapes()
        {
            $this->builder = $this->getMockBuilder(Req::class);
            $this->helper($this->getMockBuilder(Req::class));
            $f = fn () => $this->getMockBuilder(Req::class);
            return $this->getMockBuilder(Mixin::class);
        }

        public function testWrongUse(): void
        {
            $mb = $this->getMockBuilder(<error descr="Abstract class: build the double with getMockForAbstractClass().">Req::class</error>);
            $mb->getMock();
            $mb->{'getMockForAbstractClass'}();
            $this->getMockBuilder(<error descr="Trait: build the double with getMockForTrait().">Mixin::class</error>);
        }

        private function helper($b) {}
    }

    $top = (new ReqTest())->getMockBuilder(<error descr="Abstract class: build the double with getMockForAbstractClass().">Req::class</error>);
}
