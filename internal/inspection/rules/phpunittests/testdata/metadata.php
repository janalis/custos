<?php
namespace Demo;
use PHPUnit\Framework\Attributes\DataProvider as Rows;
use PHPUnit\Framework\Attributes\Depends;
use PHPUnit\Framework\Attributes\Test;
use PHPUnit\Framework\Attributes\CoversClass;
class Covered {}
#[CoversClass(Covered::class)]
class MetadataTest {
    #[Test]
    public function produces() {}
    public function rows() { return ['one' => [1]]; }
    #[Rows('rows')]
    #[Depends(methodName: 'produces')]
    public function valid() {}
    /** @depends clone produces */
    public function cloned() {}
    /** @depends shallowClone produces */
    public function shallow() {}
    #[Rows('')]
    public function <error descr="The @dataProvider target cannot be resolved to a method.">emptyProvider</error>() {}
    #[Depends('not-a-method')]
    public function <error descr="The @depends target is missing or is not a test.">invalidDependency</error>() {}
    #[Rows('missing')]
    public function <error descr="The @dataProvider target cannot be resolved to a method.">missingProvider</error>() {}
    #[Depends('missing')]
    public function <error descr="The @depends target is missing or is not a test.">missingDependency</error>() {}
    #[Depends('rows')]
    public function <error descr="The @depends target is missing or is not a test.">notTest</error>() {}
    #[CoversClass(className: 'Demo\Covered')]
    public function coveredString() {}
    #[CoversClass('UnknownCovered')]
    public function <error descr="The @covers target '\UnknownCovered' cannot be resolved.">missingCovered</error>() {}
    #[Rows($dynamic)]
    #[Depends(methodName: $dynamic)]
    #[Rows('rows', methodName: 'missing')]
    #[Rows(...$args)]
    #[Rows()]
    #[Rows(1)]
    #[Rows('rows', ...$args)]
    #[Rows(methodName: 'rows', methodName: 'missing')]
    #[Rows('rows', bogus: 1, bogus: 2)]
    #[Unrelated\DataProvider('missing')]
    #[DataProvider('missing')]
    #[CoversClass(parent::class)]
    #[CoversClass($dynamic)]
    #[CoversClass(Covered::SOMETHING)]
    #[CoversClass($dynamic::class)]
    public function ignored() {}
    #[Test]
    public function testAlreadyNamed() {}
}
#[CoversClass(UnknownCovered::class)]
class <error descr="The @covers target '\Demo\UnknownCovered' cannot be resolved.">MissingClassTarget</error> {}

class UnindexedParentTest extends ExternalParent {
    #[Rows('possiblyInherited')]
    #[Depends('possiblyInherited')]
    public function ignoredUnresolvedHierarchy() {}
}
