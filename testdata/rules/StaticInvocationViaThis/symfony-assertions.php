<?php
namespace Symfony\Bundle\FrameworkBundle\Test {
    trait WebTestAssertionsTrait {
        public static function assertResponseIsSuccessful(string $message = ''): void {}
    }
    abstract class WebTestCase {
        use WebTestAssertionsTrait;
    }
}

namespace App\Tests {
    use Symfony\Bundle\FrameworkBundle\Test\WebTestCase;

    class HomeTest extends WebTestCase {
        public static function assertHomeRendered(): void {}

        public function testHome(): void {
            // Symfony's assertion traits follow PHPUnit's $this-> convention
            $this->assertResponseIsSuccessful();
            <warning descr="Static method assertHomeRendered() called through $this; use self::assertHomeRendered().">$this</warning>->assertHomeRendered();
        }
    }
}
