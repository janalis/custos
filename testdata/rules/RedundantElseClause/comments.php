<?php
function f($a) {
    if ($a) {
        return 1;
    } /* keep me */ <warning descr="Drop the 'else' and move its code after the 'if'.">else</warning> {
        work();
    }
    if ($a > 1) {
        return 2;
    } // note
    <warning descr="Turn this 'elseif' into a separate 'if'.">elseif</warning> ($a > 0) {
        work();
    }
}
