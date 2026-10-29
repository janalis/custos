<?php
namespace App\Tests {
    use PHPUnit\Framework\TestCase;

    abstract class ApiTestCase extends TestCase
    {
        protected function assertPayload(mixed $data): void {}
    }

    final class AutocompleteTest extends ApiTestCase
    {
        public function testResults(string $body): void
        {
            $data = json_decode($body, true);
            static::assertArrayHasKey('results', $data);
            $page = json_decode($body, true);
            $this->assertIsArray($page);
            $doc = json_decode($body, true);
            parent::assertCount(2, $doc);
            $own = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($body, true)</weak_warning>;
            $this->assertPayload($own);
            $other = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($body, true)</weak_warning>;
            $this->client->assertIsArray($other);
            $named = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($body, true)</weak_warning>;
            Assert::assertIsArray($named);
            $copy = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($body, true)</weak_warning>;
            $that = $this;
            $that->assertIsArray($copy);
        }
    }

    final class Root
    {
        public function check(string $body): void
        {
            $data = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($body, true)</weak_warning>;
            $this->assertIsArray($data);
        }
    }

    final class Plain extends \App\Base
    {
        public function check(string $body): void
        {
            $data = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($body, true)</weak_warning>;
            $this->assertIsArray($data);
        }
    }
}

namespace {
    $check = function (string $body) {
        $data = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($body, true)</weak_warning>;
        $this->assertIsArray($data);
    };
}
