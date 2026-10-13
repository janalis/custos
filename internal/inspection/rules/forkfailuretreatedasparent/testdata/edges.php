<?php
if($unknown){}
function guardedFork(){ $pid=pcntl_fork();if($pid===-1){return;}if($pid){echo "parent";} }
