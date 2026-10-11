<?php
$q=msg_get_queue(123);if(msg_send($q,3,"hello",false)){msg_receive($q,3,$type,128,$message,false);}msg_receive($q,$unknown,$type,128,$message,true);msg_receive($q,0,$type,128,$message,true);function dead(){return;msg_receive($q,3,$type,128,$message,true);}

strlen("unrelated");
$q=msg_get_queue(10);msg_receive($q,3,$t,128,$m,true);
$q=msg_get_queue(11);msg_send($q,3,'hello',false);msg_receive($q,3,$t,128,$m,true);
$q=msg_get_queue(12);if(msg_send($q,4,'hello',false)){msg_receive($q,3,$t,128,$m,true);}
$q=msg_get_queue(13);if(msg_send($q,3,'hello',false)){msg_receive($q,3,$t,128,$m,true,MSG_EXCEPT);}
$q=msg_get_queue(20);if(msg_send($q,3,'hello',false)){$alias=&$q;unknown($alias);msg_receive($q,3,$t,128,$m,true);}

function steppedSendFlag() { $q = msg_get_queue(30); $s = 0; $s++; if (msg_send($q, 3, 'hello', $s)) { msg_receive($q, 3, $t, 128, $m, true); } }
function steppedReceiveFlag() { $q = msg_get_queue(31); $u = 1; --$u; if (msg_send($q, 3, 'hello', false)) { msg_receive($q, 3, $t, 128, $m, ($u)); } }
function steppedSendType() { $q = msg_get_queue(32); $kind = 3; $kind--; $want = 3; if (msg_send($q, $kind, 'hello', false)) { msg_receive($q, $want, $t, 128, $m, true); } }
function steppedReceiveType() { $q = msg_get_queue(33); $want = 3; $want++; if (msg_send($q, 3, 'hello', false)) { msg_receive($q, $want, $t, 128, $m, true); } }
function steppedPayload() { $q = msg_get_queue(34); $body = 'hello'; $body++; if (msg_send($q, 3, $body, false)) { msg_receive($q, 3, $t, 128, $m, true); } }
function capturedFlag() { $q = msg_get_queue(35); $s = false; $set = function () use (&$s) { $s = true; }; $set(); if (msg_send($q, 3, 'hello', $s)) { msg_receive($q, 3, $t, 128, $m, true); } }
function loopedFlag() { $q = msg_get_queue(36); $s = 0; foreach ([1, 2] as $round) { if ($round > 0) { if (msg_send($q, 3, 'hello', $s)) { msg_receive($q, 3, $t, 128, $m, true); } } $s++; } }
