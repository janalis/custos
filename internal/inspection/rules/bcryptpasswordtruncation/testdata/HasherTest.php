<?php
namespace App\Security;

final class HasherTest
{
    public function testLongPassword(): void
    {
        $hash = password_hash(str_repeat('a', 100), PASSWORD_BCRYPT);
    }
}
