<?php
$sem=sem_get(123);sem_acquire($sem,true);<warning descr="Check semaphore acquisition before shared-state mutation.">shm_put_var($memory,3,"value")</warning>;sem_release($sem);
