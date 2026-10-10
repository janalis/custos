<?php
$c=deflate_init(ZLIB_ENCODING_GZIP); $out=deflate_add($c,"sample",ZLIB_SYNC_FLUSH); <warning descr="Finish the compressed stream before publishing it.">file_put_contents("result.gz",$out)</warning>;
