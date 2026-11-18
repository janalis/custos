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
            $own = json_decode($body, true, 512, JSON_THROW_ON_ERROR);
            $this->assertPayload($own);
            $other = json_decode($body, true, 512, JSON_THROW_ON_ERROR);
            $this->client->assertIsArray($other);
            $named = json_decode($body, true, 512, JSON_THROW_ON_ERROR);
            Assert::assertIsArray($named);
            $copy = json_decode($body, true, 512, JSON_THROW_ON_ERROR);
            $that = $this;
            $that->assertIsArray($copy);
        }
    }

    final class Root
    {
        public function check(string $body): void
        {
            $data = json_decode($body, true, 512, JSON_THROW_ON_ERROR);
            $this->assertIsArray($data);
        }
    }

    final class Plain extends \App\Base
    {
        public function check(string $body): void
        {
            $data = json_decode($body, true, 512, JSON_THROW_ON_ERROR);
            $this->assertIsArray($data);
        }
    }
}

namespace {
    $check = function (string $body) {
        $data = json_decode($body, true, 512, JSON_THROW_ON_ERROR);
        $this->assertIsArray($data);
    };
}
