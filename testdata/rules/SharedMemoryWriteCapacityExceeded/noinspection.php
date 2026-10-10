<?php
// @noinspection SharedMemoryWriteCapacityExceeded

$s=shmop_open(123,"n",0600,20);shmop_write($s,"abcdef",shmop_size($s)-2);shmop_write($s,"abc",19);
