<?php
namespace PHPUnit\Framework {
    abstract class Assert
    {
        public static function assertIsArray(mixed $actual, string $message = ""): void {}
        public static function assertInstanceOf(string $expected, mixed $actual, string $message = ""): void {}
        public static function assertNotNull(mixed $actual, string $message = ""): void {}
        public static function assertNotFalse(mixed $actual, string $message = ""): void {}
        public static function assertSame(mixed $expected, mixed $actual, string $message = ""): void {}
        public static function assertArrayHasKey(mixed $key, array $array, string $message = ""): void {}
        public static function assertCount(int $expectedCount, iterable $haystack, string $message = ""): void {}
        public static function assertNotEmpty(mixed $actual, string $message = ""): void {}
        public static function assertEmpty(mixed $actual, string $message = ""): void {}
        public static function assertEquals(mixed $expected, mixed $actual, string $message = ""): void {}
    }
    abstract class TestCase extends Assert {}
}

namespace App\Tests {
    use PHPUnit\Framework\TestCase;

    final class DecodedTest extends TestCase
    {
        public function testChecked(string $body, array $q): void
        {
            $data = json_decode($body, true);
            static::assertIsArray($data, 'Response should be valid JSON');
            $doc = json_decode($body);
            $this->assertInstanceOf(\stdClass::class, $doc);
            $raw = json_decode($body);
            self::assertNotNull($raw);
            $json = json_encode($q);
            $this->assertNotFalse($json);
            $page = json_decode($body, true);
            $this->assertArrayHasKey('results', $page);
            $list = json_decode($body, true);
            self::assertCount(2, $list);
            $text = json_encode($q);
            $this->assertNotEmpty($text);
            $this->assertSame(['a' => 1], json_decode($body, true));
            $this->assertSame(json_decode($body), 'x');
            self::assertSame(false, json_decode($body));
            self::assertSame(null, json_encode($q));
            $this->assertSame(true, json_decode($body));
            $this->assertEquals(['x'], json_decode($body, true));
            $this->assertEquals('{"a":1}', json_encode($q));
            $this->assertEquals(2.5, json_decode($body));
            $expected = ['warnings' => ['a', 'b']];
            $this->assertEquals($expected, json_decode($body, true));
            $name = $q ? 'x' : 'y';
            $this->assertSame($name, json_decode($body));
        }

        public function testUnchecked(string $body, array $q): void
        {
            $data = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($body, true)</weak_warning>;
            $this->assertSame(null, $data);
            $list = json_decode($body, true);
            $this->assertSame([], $list);
            $json = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($q)</weak_warning>;
            $this->assertNotNull($json);
            $raw = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($body)</weak_warning>;
            $this->assertNotFalse($raw);
            $doc = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($body)</weak_warning>;
            $this->assertInstanceOf($doc, \stdClass::class);
            $named = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($body)</weak_warning>;
            $this->assertIsArray(actual: $named);
            $own = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($body)</weak_warning>;
            assertIsArray($own);
            $none = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($body)</weak_warning>;
            $this->assertEmpty($none);
            $key = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($body)</weak_warning>;
            $this->assertArrayHasKey($key, []);
            $this->assertSame(null, <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($body)</weak_warning>);
            $this->assertSame(<weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($q)</weak_warning>, false);
            $this->assertSame($q, <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($body)</weak_warning>);
            $this->assertEquals([], <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($body, true)</weak_warning>);
            $this->assertEquals('0', <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($body)</weak_warning>);
            $this->assertEquals('', <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($q)</weak_warning>);
            $this->assertEquals(0, <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($body)</weak_warning>);
            $this->assertEquals(0x1, <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($body)</weak_warning>);
            $this->assertEquals(false, <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($q)</weak_warning>);
            $this->assertEquals(PHP_EOL, <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($q)</weak_warning>);
            $empty = [];
            $this->assertEquals($empty, <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($body, true)</weak_warning>);
            $grown = ['a'];
            unset($grown[0]);
            $this->assertEquals($grown, <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($body, true)</weak_warning>);
            $copy = $empty;
            $this->assertEquals($copy, <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($body, true)</weak_warning>);
            $popped = ['a'];
            array_pop($popped);
            $this->assertEquals($popped, <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($body, true)</weak_warning>);
            $this->assertSame(expected: [], actual: <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($body)</weak_warning>);
            $this->assertSame(...[[], <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($body)</weak_warning>]);
            $this->assertSame([], 'x', <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($body)</weak_warning>);
            $this->assertSame(<weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($body)</weak_warning>);
        }
    }
}
