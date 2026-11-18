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
            $this->assertSame($flag, json_encode($locales, JSON_THROW_ON_ERROR));
            $this->assertSame($maybe, json_decode($text, true, 512, JSON_THROW_ON_ERROR));
            $this->assertSame($any, json_decode($text, true, 512, JSON_THROW_ON_ERROR));
            $this->assertSame($loose, json_decode($text, true, 512, JSON_THROW_ON_ERROR));
            $this->assertEquals($locales, json_decode($text, true, 512, JSON_THROW_ON_ERROR));
            $moved = $any;
            $this->assertSame($moved, json_decode($text, true, 512, JSON_THROW_ON_ERROR));
            $this->assertSame($nulled, json_decode($text, true, 512, JSON_THROW_ON_ERROR));
            $byRef = function (string $raw) use (&$locales): void {
                $this->assertSame($locales, json_decode($raw, true, 512, JSON_THROW_ON_ERROR));
            };
            $unused = function (string $raw): void {
                $this->assertSame($nope, json_decode($raw, true, 512, JSON_THROW_ON_ERROR));
            };
            $this->assertSame($flag, json_decode($text));
        }

        public function testCapturedExpected(string $body, $any): void
        {
            $this->assertEqualsCanonicalizing(['b', 'a'], json_decode($body, true));
            $this->assertTrue(json_decode($body));
            $this->assertTrue(json_encode($any));
            $this->assertFalse(json_decode($body, true)['saveToSentItems']);
            $this->assertFalse(json_encode($any, JSON_THROW_ON_ERROR));
            $this->assertEqualsIgnoringCase('OK', json_decode($body));
            $this->assertEqualsWithDelta(2.5, json_decode($body, false, 512, JSON_THROW_ON_ERROR), 3.0);
            $expected = ['a' => 1];
            $check = function (string $raw) use ($expected): void {
                $this->assertEqualsCanonicalizing($expected, json_decode($raw, true));
            };
            $written = function (string $raw) use ($expected): void {
                $expected = [];
                $this->assertEquals($expected, json_decode($raw, true, 512, JSON_THROW_ON_ERROR));
            };
            $grown = function (string $raw) use ($expected): void {
                $expected[] = 2;
                $this->assertEquals($expected, json_decode($raw, true, 512, JSON_THROW_ON_ERROR));
            };
            $popped = function (string $raw) use ($expected): void {
                array_pop($expected);
                $this->assertEquals($expected, json_decode($raw, true, 512, JSON_THROW_ON_ERROR));
            };
            $moved = $any;
            $unknown = function (string $raw) use ($moved): void {
                $this->assertEquals($moved, json_decode($raw, true, 512, JSON_THROW_ON_ERROR));
            };
            $other = function (string $raw) use ($any): void {
                $this->assertEquals($expected, json_decode($raw, true, 512, JSON_THROW_ON_ERROR));
            };
        }

        public function testUnchecked(string $body, array $q): void
        {
            $data = json_decode($body, true, 512, JSON_THROW_ON_ERROR);
            $this->assertSame(null, $data);
            $list = json_decode($body, true);
            $this->assertSame([], $list);
            $json = json_encode($q, JSON_THROW_ON_ERROR);
            $this->assertNotNull($json);
            $raw = json_decode($body, false, 512, JSON_THROW_ON_ERROR);
            $this->assertNotFalse($raw);
            $doc = json_decode($body, false, 512, JSON_THROW_ON_ERROR);
            $this->assertInstanceOf($doc, \stdClass::class);
            $named = json_decode($body, false, 512, JSON_THROW_ON_ERROR);
            $this->assertIsArray(actual: $named);
            $own = json_decode($body, false, 512, JSON_THROW_ON_ERROR);
            assertIsArray($own);
            $none = json_decode($body, false, 512, JSON_THROW_ON_ERROR);
            $this->assertEmpty($none);
            $key = json_decode($body, false, 512, JSON_THROW_ON_ERROR);
            $this->assertArrayHasKey($key, []);
            $this->assertSame(null, json_decode($body, false, 512, JSON_THROW_ON_ERROR));
            $this->assertSame(json_encode($q, JSON_THROW_ON_ERROR), false);
            $this->assertSame($q, json_decode($body));
            $this->assertEquals([], json_decode($body, true, 512, JSON_THROW_ON_ERROR));
            $this->assertEquals('0', json_decode($body, false, 512, JSON_THROW_ON_ERROR));
            $this->assertEquals('', json_encode($q, JSON_THROW_ON_ERROR));
            $this->assertEquals(0, json_decode($body, false, 512, JSON_THROW_ON_ERROR));
            $this->assertEquals(0x1, json_decode($body, false, 512, JSON_THROW_ON_ERROR));
            $this->assertEquals(false, json_encode($q, JSON_THROW_ON_ERROR));
            $this->assertEquals(PHP_EOL, json_encode($q, JSON_THROW_ON_ERROR));
            $empty = [];
            $this->assertEquals($empty, json_decode($body, true, 512, JSON_THROW_ON_ERROR));
            $grown = ['a'];
            unset($grown[0]);
            $this->assertEquals($grown, json_decode($body, true, 512, JSON_THROW_ON_ERROR));
            $copy = $empty;
            $this->assertEquals($copy, json_decode($body, true, 512, JSON_THROW_ON_ERROR));
            $popped = ['a'];
            array_pop($popped);
            $this->assertEquals($popped, json_decode($body, true, 512, JSON_THROW_ON_ERROR));
            $this->assertSame(expected: [], actual: json_decode($body, false, 512, JSON_THROW_ON_ERROR));
            $this->assertSame(...[[], json_decode($body, false, 512, JSON_THROW_ON_ERROR)]);
            $this->assertSame([], 'x', json_decode($body, false, 512, JSON_THROW_ON_ERROR));
            $this->assertSame(json_decode($body, false, 512, JSON_THROW_ON_ERROR));
            $this->assertSame(null, json_decode($body, true, 512, JSON_THROW_ON_ERROR)['data']);
            $this->assertEmpty(json_decode($body, false, 512, JSON_THROW_ON_ERROR)->items);
            $this->assertStringContainsString('ok', json_encode($q, JSON_THROW_ON_ERROR));
            $whole = json_decode($body, false, 512, JSON_THROW_ON_ERROR);
            $copy = $whole;
            $this->assertNotNull($copy);
            $tree = json_decode($body, true, 512, JSON_THROW_ON_ERROR);
            $items = $tree['items'];
            $this->assertSame(null, $items);
            $node = json_decode($body, false, 512, JSON_THROW_ON_ERROR);
            $parts = [$node->a];
            $this->assertNotNull($parts);
            $leaf = json_decode($body, false, 512, JSON_THROW_ON_ERROR);
            $leaf->x .= 'y';
            $label = json_decode($body, false, 512, JSON_THROW_ON_ERROR);
            $text = $label->name;
        }
    }
}
