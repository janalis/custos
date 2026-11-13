<?php
// At the newest PHP version, no later one may change a builtin.
function latest(array $list): bool
{
    return <warning descr="This check is always false: the value is true.">sort($list) === null</warning>;
}
