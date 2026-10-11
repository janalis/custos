<?php
function laterStep() { $q = msg_get_queue(40); $s = 0; if (msg_send($q, 3, 'hello', $s)) { <warning descr="Use matching serialization flags for the message.">msg_receive($q, 3, $t, 128, $m, true)</warning>; } $s++; }
function valueCapture() { $q = msg_get_queue(41); $s = false; $show = function () use ($s) { return $s; }; if (msg_send($q, 3, 'hello', $s)) { <warning descr="Use matching serialization flags for the message.">msg_receive($q, 3, $t, 128, $m, true)</warning>; } }
function otherLoop() { $q = msg_get_queue(42); $s = 0; foreach ([1, 2] as $round) { if ($round > 0) { if (msg_send($q, 3, 'hello', $s)) { <warning descr="Use matching serialization flags for the message.">msg_receive($q, 3, $t, 128, $m, true)</warning>; } } } $s--; }
