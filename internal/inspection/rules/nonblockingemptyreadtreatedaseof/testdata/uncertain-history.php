<?php
$h=fopen('php://memory','r+');if($flag){stream_set_blocking($h,false);}if(fread($h,1)===''){fclose($h);}
