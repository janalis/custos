<?php

final class Money
{
    /** @return string */
    public function __TOSTRING() { return 'x'; }

    /** @return array */
    public function __DebugInfo() { return []; }

    /** @return string */
    public function <weak_warning descr="Declare ': string' as the return type.">label</weak_warning>() { return 'x'; }
}
