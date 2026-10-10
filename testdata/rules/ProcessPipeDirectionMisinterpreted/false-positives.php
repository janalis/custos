<?php
$p=proc_open(["cat"],[0=>["pipe","r"]],$pipes);if(is_resource($p)){fwrite($pipes[0],"input");}fwrite($unknown,"x");function dead(){return;fread($pipes[0],5);}

strlen("unrelated");

fwrite($pipes[0],"x");$p=proc_open("cat",$descriptors,$pipes);if(is_resource($p)){fwrite($pipes[0],"x");}$p=proc_open("cat",[0=>$tuple],$pipes);if(is_resource($p)){fwrite($pipes[0],"x");}$p=proc_open("cat",[0=>["file","path"]],$pipes);if(is_resource($p)){fwrite($pipes[$unknown],"x");}

$p=proc_open('cat',[0=>['pipe','w']],$pipes);if(is_resource($p)){$alias=&$pipes;unknown($alias);fwrite($pipes[0],'x');}
