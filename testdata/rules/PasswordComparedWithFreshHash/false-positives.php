<?php
namespace Local; function password_hash($value, $algo) { return $value; } if (password_hash($password, 1) === $stored) { return true; }
namespace Checks;
function store($password)
{
    if (\password_hash($password, PASSWORD_BCRYPT) === false) { throw new \RuntimeException('hashing failed'); }
    if (\password_hash($password, PASSWORD_DEFAULT) === null) { throw new \RuntimeException('hashing failed'); }
    if (TRUE !== \password_hash($password, PASSWORD_DEFAULT)) { return null; }
    $hash = \password_hash($password, PASSWORD_DEFAULT);
    if ($hash === \false) { throw new \RuntimeException('hashing failed'); }
    if ($hash === 12) { return null; }
    if ($hash == 1.5) { return null; }
    if ($hash === []) { return null; }
    if ($hash == null || $hash !== '') { return $hash; }
    $failure = false;
    if ($hash === $failure) { return null; }
    return $hash;
}
