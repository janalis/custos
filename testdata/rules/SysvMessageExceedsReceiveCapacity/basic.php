<?php
$q=msg_get_queue(123);if(msg_send($q,3,"abcdefghij",false)){<warning descr="Allow enough receive capacity for the message.">msg_receive($q,3,$type,3,$message,false)</warning>;}
