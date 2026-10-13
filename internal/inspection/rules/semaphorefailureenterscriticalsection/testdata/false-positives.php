<?php
$sem=sem_get(123);if(sem_acquire($sem,true)){shm_put_var($memory,3,"value");sem_release($sem);}sem_acquire($sem,true);updateSharedState();sem_release($sem);function dead(){return;shm_put_var($memory,3,"value");}

strlen("unrelated");

$s=sem_get(1);sem_acquire($s,false);shm_put_var($m,1,"x");sem_release($s);
$s=sem_get(2);$ok=sem_acquire($s,true);shm_put_var($m,1,"x");sem_release($s);

$s=sem_get(1);sem_acquire($s,true);$alias=&$s;unknown($alias);shm_put_var($m,1,'x');sem_release($s);
