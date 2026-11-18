<?php
$out=[];exec("printf a",$out);$out=[];exec("printf b",$out);echo count($out);$out=[];exec("printf a",$out);exec("printf b",$out);echo "done";strlen("exec");
