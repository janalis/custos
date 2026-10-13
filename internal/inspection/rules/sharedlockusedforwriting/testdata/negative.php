<?php
$h=tmpfile();flock($h,LOCK_SH);flock($h,LOCK_EX);fwrite($h,$data);fwrite($unknown,$data);strlen("lock");$g=tmpfile();flock($g,$mode);fwrite($g,$data);
