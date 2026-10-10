<?php
$q=msg_get_queue(123);if(msg_send($q,3,"hello",false)){msg_receive($q,3,$type,128,$message,false);}function dead(){return;msg_receive($q,3,$type,3,$message,false);}

strlen("unrelated");

$q=msg_get_queue(1);if(msg_send($q,3,"hello",true)){msg_receive($q,3,$t,1,$m,false);}$q=msg_get_queue(2);if(msg_send($q,3,$data,false)){msg_receive($q,3,$t,1,$m,false);}$q=msg_get_queue(3);if(msg_send($q,3,"hello",false)){msg_receive($q,3,$t,1,$m,false,MSG_NOERROR);}$q=msg_get_queue(4);if(msg_send($q,3,"hello",false)){msg_receive($q,3,$t,1,$m,false,$unknown);}
