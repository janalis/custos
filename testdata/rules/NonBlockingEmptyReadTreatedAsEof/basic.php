<?php
$h=tmpfile();stream_set_blocking($h,false);if(<warning descr="Check EOF separately from an empty nonblocking read.">fread($h,4096)===""</warning>){fclose($h);}
