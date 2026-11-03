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
        public static function assertTrue(mixed $condition, string $message = ""): void {}
        public static function assertFalse(mixed $condition, string $message = ""): void {}
        public static function assertEqualsCanonicalizing(mixed $expected, mixed $actual, string $message = ""): void {}
        public static function assertEqualsIgnoringCase(mixed $expected, mixed $actual, string $message = ""): void {}
        public static function assertEqualsWithDelta(mixed $expected, mixed $actual, float $delta, string $message = ""): void {}
        public static function assertStringContainsString(string $needle, string $haystack, string $message = ""): void {}
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
            $this->assertSame([], json_decode($body, true)['data']['clients']);
            self::assertNotNull(json_decode($body)->id);
            $result = json_decode($body, true);
            $html = $result['results'][0]['html'];
            $this->assertStringContainsString('<div>', $html);
            $page = json_decode($body);
            $status = $page->status;
            $this->assertSame('ok', $status);
            $this->assertStringContainsString('ok', json_decode($body));
        }

        public function testTypedExpected(array $locales, string $text, bool $flag, ?array $maybe, mixed $any, $loose, array $moved, array $nulled = null, string ...$rest): void
        {
            $this->assertSame($locales, json_decode($text, true));
            $this->assertSame($text, json_encode($locales));
            $check = function (string $raw) use ($locales): void {
                $this->assertSame($locales, json_decode($raw, true)['langs']);
            };
            $arrow = fn (string $raw) => $this->assertSame($locales, json_decode($raw, true));
            $this->assertSame($flag, <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($locales)</weak_warning>);
            $this->assertSame($maybe, <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($text, true)</weak_warning>);
            $this->assertSame($any, <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($text, true)</weak_warning>);
            $this->assertSame($loose, <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($text, true)</weak_warning>);
            $this->assertEquals($locales, <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($text, true)</weak_warning>);
            $moved = $any;
            $this->assertSame($moved, <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($text, true)</weak_warning>);
            $this->assertSame($nulled, <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($text, true)</weak_warning>);
            $byRef = function (string $raw) use (&$locales): void {
                $this->assertSame($locales, <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($raw, true)</weak_warning>);
            };
            $unused = function (string $raw): void {
                $this->assertSame($nope, <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($raw, true)</weak_warning>);
            };
            $this->assertSame($flag, json_decode($text));
        }

        public function testCapturedExpected(string $body, $any): void
        {
            $this->assertEqualsCanonicalizing(['b', 'a'], json_decode($body, true));
            $this->assertTrue(json_decode($body));
            $this->assertTrue(json_encode($any));
            $this->assertFalse(json_decode($body, true)['saveToSentItems']);
            $this->assertFalse(<weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($any)</weak_warning>);
            $this->assertEqualsIgnoringCase('OK', json_decode($body));
            $this->assertEqualsWithDelta(2.5, <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($body)</weak_warning>, 3.0);
            $expected = ['a' => 1];
            $check = function (string $raw) use ($expected): void {
                $this->assertEqualsCanonicalizing($expected, json_decode($raw, true));
            };
            $written = function (string $raw) use ($expected): void {
                $expected = [];
                $this->assertEquals($expected, <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($raw, true)</weak_warning>);
            };
            $grown = function (string $raw) use ($expected): void {
                $expected[] = 2;
                $this->assertEquals($expected, <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($raw, true)</weak_warning>);
            };
            $popped = function (string $raw) use ($expected): void {
                array_pop($expected);
                $this->assertEquals($expected, <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($raw, true)</weak_warning>);
            };
            $moved = $any;
            $unknown = function (string $raw) use ($moved): void {
                $this->assertEquals($moved, <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($raw, true)</weak_warning>);
            };
            $other = function (string $raw) use ($any): void {
                $this->assertEquals($expected, <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($raw, true)</weak_warning>);
            };
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
            $this->assertSame($q, json_decode($body));
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
            $this->assertSame(null, <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($body, true)</weak_warning>['data']);
            $this->assertEmpty(<weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($body)</weak_warning>->items);
            $this->assertStringContainsString('ok', <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($q)</weak_warning>);
            $whole = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($body)</weak_warning>;
            $copy = $whole;
            $this->assertNotNull($copy);
            $tree = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($body, true)</weak_warning>;
            $items = $tree['items'];
            $this->assertSame(null, $items);
            $node = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($body)</weak_warning>;
            $parts = [$node->a];
            $this->assertNotNull($parts);
            $leaf = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($body)</weak_warning>;
            $leaf->x .= 'y';
            $label = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($body)</weak_warning>;
            $text = $label->name;
        }
    }
}
