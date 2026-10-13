<?php
$h=tmpfile();stream_set_blocking($h,true);if(fread($h,4096)===""){fclose($h);}if(fread($h,4096)==="" && feof($h)){fclose($h);}if(fread($h,4096)===""){echo "empty";}if($data===""){fclose($h);}$sum=1+2;
