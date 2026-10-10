<?php
$q=msg_get_queue(random_int(100000,2000000000));try {if(msg_send($q,3,'i:7;',false)){var_dump(msg_receive($q,3,$type,128,$message,true,0,$error));var_dump($message);}} finally {msg_remove_queue($q);}
