<?php
namespace App {
    function array_push(array &$a, $v) { return 0; }
    function count($v) { return 0; }

    $queue = [];
    array_push($queue, 1);
    $queue[count($queue)] = 2;
}

namespace {
    $list = [];
    <warning descr="Use '$list[] = 1' instead; it avoids a function call.">Array_Push($list, 1)</warning>;
    $list[<warning descr="The index is redundant here; use '[]' to append.">COUNT</warning>($list)] = 2;
}
