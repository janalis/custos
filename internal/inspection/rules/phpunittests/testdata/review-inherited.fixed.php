<?php
namespace App\Tests;

use Vendor\Kit\BaseTestCase;
use Vendor\Kit\ProvidesRows;

// The parent class and the trait are not indexed: they may declare the
// members the tags name.
class InheritedTest extends BaseTestCase
{
    use ProvidesRows;

    /** @dataProvider inheritedRows */
    public function testA($r) {}

    /** @depends testInheritedSetup */
    public function testB() {}

    /** @covers InheritedTest::inheritedHelper */
    public function testC() {}

    /** @dataProvider \App\Tests\InheritedTest::otherRows */
    public function testD($r) {}

    /** @dataProvider \Vendor\Kit\Rows::all */
    public function testE($r) {}

    // Nor is the class a @covers names (E9d).
    /** @covers \Vendor\Kit\Thing::run */
    public function testF() {}

    /** @dataProvider rows */
    public function testWithRows($r) {}

    /**
     * @depends testA
     */
    public function testAfterA() {}

    /** */
    public function testMail() {}

    public static function rows() { return ['one' => [1]]; }
}
