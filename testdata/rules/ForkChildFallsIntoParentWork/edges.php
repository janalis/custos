<?php
if($other===0){echo "x";}echo "tail";$p=pcntl_fork();if($p==0){echo "x";}echo "tail";if($p===1){echo "x";}echo "tail";if($p===0){}echo "tail";if($p===0)echo "x";echo "tail";if($p===0){return;}echo "tail";if($p===0){echo "x";}else{echo "parent";}echo "tail";
$p=pcntl_fork();if($p===0){throw new RuntimeException();}echo "tail";function lastChild(){$p=pcntl_fork();if($p===0){echo "child";}}
function emptyChild(){$p=pcntl_fork();if($p===0){}echo "tail";}
function returnedChild(){$p=pcntl_fork();if($p===0){return;}echo "tail";}
function singleChild(){$p=pcntl_fork();if($p===0)exit;echo "tail";}
