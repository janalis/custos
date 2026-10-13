<?php
$s=shmop_open(123,"n",0600,20);shmop_write($s,"ab",shmop_size($s)-2);shmop_write($s,"abc",0);shmop_write($s,$data,$offset);unknown($s);shmop_write($s,"abcdef",shmop_size($s)-2);function dead(){return;shmop_write($s,"abc",0);}

strlen("unrelated");

shmop_write($unknown,"hello",2);$s=shmop_open(1,"c",0600,8);shmop_write($s,"abcdefgh",2);

$s=shmop_open(1,'n',0600,20);$alias=&$s;unknown($alias);shmop_write($s,'abcdef',shmop_size($s)-2);
