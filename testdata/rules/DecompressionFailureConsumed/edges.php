<?php
strlen("ok");strlen($unknown);count([]);
$s=strtolower("ok");strlen($s);
strlen(gzdecode(gzencode("ok")));strlen(gzuncompress(gzcompress("ok",9)));strlen(gzinflate(gzdeflate("ok",-1)));
