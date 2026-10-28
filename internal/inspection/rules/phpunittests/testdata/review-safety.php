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
        <weak_warning descr="Use 'assertNotTrue()' instead.">$this->assertTrue(!$ok)</weak_warning>;
        <weak_warning descr="Use 'assertNotFalse()' instead.">$this->assertFalse(!isset($a), 'm')</weak_warning>;
        <weak_warning descr="Use 'assertNotTrue()' instead.">$this->assertTrue(!$this->flag())</weak_warning>;
        <weak_warning descr="Use 'assertNotTrue()' instead.">$this->assertTrue(!empty($rows[0]))</weak_warning>;
    }

    public function testComparisons($a, int $n, float $ratio, string $s, mixed $any)
    {
        <weak_warning descr="Use 'assertEquals()' instead.">$this->assertTrue($n == 3)</weak_warning>;
        <weak_warning descr="Use 'assertEquals()' instead.">$this->assertTrue($s == '1000')</weak_warning>;
        <weak_warning descr="Use 'assertNotEquals()' instead.">$this->assertTrue($a != $n)</weak_warning>;
        <weak_warning descr="Use 'assertEquals()' instead.">$this->assertTrue($n == $a)</weak_warning>;
        <weak_warning descr="Use 'assertSame()' instead.">$this->assertTrue($a === $n)</weak_warning>;
        <weak_warning descr="Use 'assertSame()' instead.">$this->assertTrue($ratio === $a)</weak_warning>;
        <weak_warning descr="Use 'assertSame()' instead.">$this->assertFalse($any !== $ratio)</weak_warning>;
        $this->assertSame($a, $n, true);
        $this->assertNotSame($a, $n, null);
        $this->assertTrue(is_callable($a, true));
    }

    public function testClassNames($pet)
    {
        <weak_warning descr="Use 'assertInstanceOf()' instead.">$this->assertSame('Zoo\\Cage', get_class($pet))</weak_warning>;
        <weak_warning descr="Use 'assertNotInstanceOf()' instead.">$this->assertNotSame(get_class($pet), 'Zoo\Diet')</weak_warning>;
        <weak_warning descr="Use 'assertInstanceOf()' instead.">$this->assertEquals("Zoo\\Cage", get_class($pet), 'm')</weak_warning>;
        <weak_warning descr="Use 'assertNotInstanceOf()' instead.">$this->assertNotSame('Zoo\\Animal', get_class($pet))</weak_warning>;
        <weak_warning descr="Use 'assertInstanceOf()' instead.">$this->assertSame('zoo\\cage', get_class($pet))</weak_warning>;
        <weak_warning descr="Use 'assertInstanceOf()' instead.">$this->assertSame('\\Zoo\\Cage', get_class($pet))</weak_warning>;
        <weak_warning descr="Use 'assertInstanceOf()' instead.">$this->assertSame('Zoo\\Missing', get_class($pet))</weak_warning>;
        $this->assertSame("Zoo\Cage", get_class($pet));
        $this->assertNotSame('not a class', get_class($pet));
        $this->assertNotEquals('class@anonymous', get_class($pet));
        $this->assertNotSame('N/A?', get_class($pet));
    }

    public function testTypedParameters($a, int $n, string $s, array $rows, $pet)
    {
        $cage = new Cage();
        <weak_warning descr="Use 'assertCount()' instead.">$this->assertEquals($n, count($rows))</weak_warning>;
        <weak_warning descr="Use 'assertCount()' instead.">$this->assertEquals($a, count($rows))</weak_warning>;
        <weak_warning descr="Use 'assertCount()' instead.">$this->assertEquals($s, count($rows))</weak_warning>;
        <weak_warning descr="Use 'assertInstanceOf()' instead.">$this->assertTrue($pet instanceof $s)</weak_warning>;
        <weak_warning descr="Use 'assertInstanceOf()' instead.">$this->assertTrue($pet instanceof $a)</weak_warning>;
        <weak_warning descr="Use 'assertInstanceOf()' instead.">$this->assertTrue($pet instanceof $cage)</weak_warning>;
        <weak_warning descr="Use 'assertStringEqualsFile()' instead.">$this->assertSame(file_get_contents($s), $s)</weak_warning>;
        <weak_warning descr="Use 'assertStringEqualsFile()' instead.">$this->assertEquals(file_get_contents($s), $n)</weak_warning>;
    }

    public function testEmptyOperands($a, string $s, array $rows)
    {
        $local = [];
        if ($a) {
            $maybe = [];
        }
        <weak_warning descr="Use 'assertEmpty()' instead.">$this->assertTrue(empty($local))</weak_warning>;
        <weak_warning descr="Use 'assertNotEmpty()' instead.">$this->assertFalse(empty(trim($s)))</weak_warning>;
        $this->assertTrue(empty($maybe));
        $this->assertTrue(empty($rows[0]));
        $this->assertTrue(empty($s[7]));
        $this->assertTrue(empty($this->late));
    }

    private const GIF = 'data:image/gif;base64,R0lGODdhAQABAIAAAP8=';

    public function testStreams(string $s)
    {
        $this->assertSame(file_get_contents(self::GIF), $s);
        $this->assertSame(file_get_contents('PHP://memory'), $s);
        $this->assertSame(file_get_contents('php://memory'), file_get_contents('/tmp/x'));
        <weak_warning descr="Use 'assertStringEqualsFile()' instead.">$this->assertSame(file_get_contents('file:///tmp/x'), $s)</weak_warning>;
        <weak_warning descr="Use 'assertStringEqualsFile()' instead.">$this->assertSame(file_get_contents('vfs://root/a'), $s)</weak_warning>;
        $this->assertSame(file_get_contents('https://example.com/a'), $s);
    }
}
