<?php
inflate_add($unknown,"data");$c=inflate_init(ZLIB_ENCODING_RAW);inflate_add($c,$unknown);$d=strtolower("x");inflate_add($c,$d);
$c=inflate_init(ZLIB_ENCODING_RAW);inflate_add($c,gzencode("x",-1,FORCE_DEFLATE));
