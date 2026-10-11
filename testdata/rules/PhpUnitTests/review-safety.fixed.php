<?php
namespace Zoo;

final class Cage {}
class Animal {}
enum Diet { case Meat; }

class SafetyTest
{
    public string $late;

    public function flag(): int { return 1; }

    public function testInverted($a, bool $ok, array $rows)
    {
        $this->assertNotTrue($ok);
        $this->assertNotFalse(isset($a), 'm');
        $this->assertTrue(!$this->flag());
        $this->assertNotTrue(empty($rows[0]));
    }

    public function testComparisons($a, int $n, float $ratio, string $s, mixed $any)
    {
        $this->assertEquals($n, 3);
        $this->assertTrue($s == '1000');
        $this->assertTrue($a != $n);
        $this->assertTrue($n == $a);
        $this->assertSame($a, $n);
        $this->assertTrue($ratio === $a);
        $this->assertFalse($any !== $ratio);
        $this->assertSame($a, $n, true);
        $this->assertNotSame($a, $n, null);
        $this->assertTrue(is_callable($a, true));
    }

    public function testClassNames($pet)
    {
        $this->assertInstanceOf(\Zoo\Cage::class, $pet);
        $this->assertNotInstanceOf(\Zoo\Diet::class, $pet);
        $this->assertInstanceOf(\Zoo\Cage::class, $pet, 'm');
        $this->assertNotSame('Zoo\\Animal', get_class($pet));
        $this->assertSame('zoo\\cage', get_class($pet));
        $this->assertSame('\\Zoo\\Cage', get_class($pet));
        $this->assertSame('Zoo\\Missing', get_class($pet));
        $this->assertSame("Zoo\Cage", get_class($pet));
        $this->assertNotSame('not a class', get_class($pet));
        $this->assertNotEquals('class@anonymous', get_class($pet));
        $this->assertNotSame('N/A?', get_class($pet));
    }

    public function testTypedParameters($a, int $n, string $s, array $rows, $pet)
    {
        $cage = new Cage();
        $this->assertCount($n, $rows);
        $this->assertCount($a, $rows);
        $this->assertEquals($s, count($rows));
        $this->assertInstanceOf($s, $pet);
        $this->assertTrue($pet instanceof $a);
        $this->assertTrue($pet instanceof $cage);
        $this->assertStringEqualsFile($s, $s);
        $this->assertEquals(file_get_contents($s), $n);
    }

    public function testEmptyOperands($a, string $s, array $rows)
    {
        $local = [];
        if ($a) {
            $maybe = [];
        }
        $this->assertEmpty($local);
        $this->assertNotEmpty(trim($s));
        $this->assertTrue(empty($maybe));
        $this->assertTrue(empty($rows[0]));
        $this->assertTrue(empty($s[7]));
        $this->assertTrue(empty($this->late));
    }
}
