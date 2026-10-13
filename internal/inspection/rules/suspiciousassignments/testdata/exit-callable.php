<?php
function exitCallableSwitch($kind) {
    switch ($kind) {
        case 1:
            $value = 'first';
            exit(...);
        case 2:
            <error descr="This write overwrites a value set in a previous case; a 'break' may be missing.">$value</error> = 'second';
            break;
    }
    return $value;
}
function dieCallableSwitch($kind) {
    switch ($kind) {
        case 1:
            $value = 'first';
            die(...);
        case 2:
            <error descr="This write overwrites a value set in a previous case; a 'break' may be missing.">$value</error> = 'second';
            break;
    }
    return $value;
}
function immediateExitInvocationSwitch($kind) {
    switch ($kind) {
        case 1:
            $value = 'first';
            ((exit(...))(1));
        case 2:
            $value = 'second';
            break;
    }
    return $value;
}
