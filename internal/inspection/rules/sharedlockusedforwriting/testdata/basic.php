<?php
$h=tmpfile();flock($h,LOCK_SH);<warning descr="Acquire an exclusive lock before writing.">fwrite($h,$data)</warning>;
$g=tmpfile();flock($g,LOCK_NB|LOCK_SH);<warning descr="Acquire an exclusive lock before writing.">ftruncate($g,0)</warning>;
