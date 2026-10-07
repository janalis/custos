<?php

final class Money
{
    /** @return string */
    public function __TOSTRING() { return 'x'; }

    /** @return array */
    public function __DebugInfo() { return []; }

    /** @return string */
    public function label(): string { return 'x'; }
}
