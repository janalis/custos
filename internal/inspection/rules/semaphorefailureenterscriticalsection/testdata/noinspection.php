<?php
// @noinspection SemaphoreFailureEntersCriticalSection

$sem=sem_get(123);sem_acquire($sem,true);shm_put_var($memory,3,"value");sem_release($sem);
