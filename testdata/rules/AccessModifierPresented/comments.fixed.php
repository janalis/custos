<?php

class Registry
{
    // A comment between modifiers would be lost by the rewrite: report only.
    static /* kept for BC, see #123 */ final function make() { return 'make'; }
    static // shared
        readonly int $count;

    // Comments outside the modifier list do not get in the way.
    /** Builds. */
    public static /* after */ function build() {}
    /* before */ final public static function seal() {}
    public function /* anonymous? */ named() {}
}
