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
    $list[] = 1;
    $list[COUNT($list)] = 2;
}
