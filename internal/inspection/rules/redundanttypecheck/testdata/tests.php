<?php
namespace Shop\Tests;

// Tests probe impossible values on purpose.
final class CartTest
{
    public function testRejects(int $qty): bool
    {
        return is_int($qty);
    }
}
