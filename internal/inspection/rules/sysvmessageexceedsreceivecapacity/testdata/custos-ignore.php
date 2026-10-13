<?php
// @custos-ignore SysvMessageExceedsReceiveCapacity

$q=msg_get_queue(123);if(msg_send($q,3,"abcdefghij",false)){msg_receive($q,3,$type,3,$message,false);}
