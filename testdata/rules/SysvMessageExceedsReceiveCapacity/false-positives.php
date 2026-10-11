<?php
$q=msg_get_queue(123);if(msg_send($q,3,"hello",false)){msg_receive($q,3,$type,128,$message,false);}function dead(){return;msg_receive($q,3,$type,3,$message,false);}

strlen("unrelated");

$q=msg_get_queue(1);if(msg_send($q,3,"hello",true)){msg_receive($q,3,$t,1,$m,false);}$q=msg_get_queue(2);if(msg_send($q,3,$data,false)){msg_receive($q,3,$t,1,$m,false);}$q=msg_get_queue(3);if(msg_send($q,3,"hello",false)){msg_receive($q,3,$t,1,$m,false,MSG_NOERROR);}$q=msg_get_queue(4);if(msg_send($q,3,"hello",false)){msg_receive($q,3,$t,1,$m,false,$unknown);}

// The capacity is stepped after its literal assignment: it is 10 at the receive.
function steppedCapacity()
{
    $size = 9;
    $size++;
    $q = msg_get_queue(201);
    if (msg_send($q, 3, "abcdefghij", false)) {
        msg_receive($q, 3, $type, $size, $message, false);
    }
}

function preDecrementedCapacity()
{
    $size = 11;
    --$size;
    $q = msg_get_queue(202);
    if (msg_send($q, 3, "abcdefghij", false)) {
        msg_receive($q, 3, $type, ($size), $message, false);
    }
}

// A by-reference capture may rewrite the capacity before the receive.
function capturedCapacity()
{
    $size = 3;
    $grow = function () use (&$size) { $size = 64; };
    $grow();
    $q = msg_get_queue(203);
    if (msg_send($q, 3, "abcdefghij", false)) {
        msg_receive($q, 3, $type, $size, $message, false);
    }
}

// A step later in the same loop reaches the receive on the next iteration.
function loopCapacity(array $items)
{
    $size = 3;
    $q = msg_get_queue(204);
    foreach ($items as $item) {
        if (msg_send($q, 3, "abcdefghij", false)) {
            msg_receive($q, 3, $type, $size, $message, false);
        }
        $size++;
    }
}

// Stepped flags, message types and sent payloads are no longer proven either.
function steppedFlags()
{
    $flags = 1;
    $flags++;
    $q = msg_get_queue(205);
    if (msg_send($q, 3, "abcdefghij", false)) {
        msg_receive($q, 3, $type, 3, $message, false, MSG_IPC_NOWAIT | $flags);
    }
}

function steppedPayload()
{
    $data = "9";
    $data--;
    $q = msg_get_queue(206);
    if (msg_send($q, 3, $data, false)) {
        msg_receive($q, 3, $type, 3, $message, false);
    }
}
