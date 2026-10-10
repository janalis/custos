<?php
$q=msg_get_queue(123);if(msg_send($q,3,"hello",false)){msg_receive($q,3,$type,128,$message,false);}msg_receive($q,$unknown,$type,128,$message,true);msg_receive($q,0,$type,128,$message,true);function dead(){return;msg_receive($q,3,$type,128,$message,true);}

strlen("unrelated");
$q=msg_get_queue(10);msg_receive($q,3,$t,128,$m,true);
$q=msg_get_queue(11);msg_send($q,3,'hello',false);msg_receive($q,3,$t,128,$m,true);
$q=msg_get_queue(12);if(msg_send($q,4,'hello',false)){msg_receive($q,3,$t,128,$m,true);}
$q=msg_get_queue(13);if(msg_send($q,3,'hello',false)){msg_receive($q,3,$t,128,$m,true,MSG_EXCEPT);}
$q=msg_get_queue(20);if(msg_send($q,3,'hello',false)){$alias=&$q;unknown($alias);msg_receive($q,3,$t,128,$m,true);}
