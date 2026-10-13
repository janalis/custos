<?php
$s=shmop_open(123,"n",0600,20);<error descr="Keep writes within shared-memory capacity.">shmop_write($s,"abcdef",shmop_size($s)-2)</error>;<error descr="Keep writes within shared-memory capacity.">shmop_write($s,"abc",19)</error>;
