<?php
declare(strict_types=1);

namespace Shop;

// PHP 8.6: the default value initializes a readonly property, so the
// constructor's write fails as a second initialization (another rule's
// concern) before its value is checked against the type.
final class Ticket
{
    public readonly int $number = 1;

    public function __construct()
    {
        $this->number = 'x';
    }
}

// Without a default, the constructor's only write initializes it.
final class Badge
{
    public readonly int $number;

    public function __construct()
    {
        $this->number = <error descr="Assign int to Shop\Badge::$number, not string.">'x'</error>;
    }
}
