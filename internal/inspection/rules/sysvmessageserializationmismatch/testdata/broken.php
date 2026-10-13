<?php
$q=msg_get_queue(123);if(msg_send($q,3,"hello",false)){msg_receive($q,3,$type,128,$message,true);}

$broken = ;
