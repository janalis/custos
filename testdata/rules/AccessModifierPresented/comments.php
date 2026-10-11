<?php

class Registry
{
    // A comment between modifiers would be lost by the rewrite: report only.
    static /* kept for BC, see #123 */ final function <weak_warning descr="Declare the visibility of 'make' explicitly.">make</weak_warning>() { return 'make'; }
    static // shared
        readonly int <weak_warning descr="Declare the visibility of 'count' explicitly.">$count</weak_warning>;

    // Comments outside the modifier list do not get in the way.
    /** Builds. */
    static /* after */ function <weak_warning descr="Declare the visibility of 'build' explicitly.">build</weak_warning>() {}
    /* before */ final static function <weak_warning descr="Declare the visibility of 'seal' explicitly.">seal</weak_warning>() {}
    function /* anonymous? */ <weak_warning descr="Declare the visibility of 'named' explicitly.">named</weak_warning>() {}
}
