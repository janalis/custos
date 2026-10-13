<?php
// A by-value capture, a plain read and a later step outside any loop keep the capacity proven.
function proven()
{
    $size = 3;
    $copy = function () use ($size) { return $size; };
    echo $size;
    $q = msg_get_queue(301);
    if (msg_send($q, 3, "abcdefghij", false)) {
        <warning descr="Allow enough receive capacity for the message.">msg_receive($q, 3, $type, $size, $message, false)</warning>;
    }
    $size++;
    return $copy;
}

// A step in a different loop after the receive does not reach it.
function laterLoop(array $items)
{
    $size = 3;
    $q = msg_get_queue(302);
    foreach ($items as $item) {
        if (msg_send($q, 3, "abcdefghij", false)) {
            <warning descr="Allow enough receive capacity for the message.">msg_receive($q, 3, $type, $size, $message, false)</warning>;
        }
        break;
    }
    while ($size < 5) {
        $size++;
    }
}
