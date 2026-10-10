<?php
// @noinspection DeflateStreamNeverFinished
$c=deflate_init(ZLIB_ENCODING_GZIP); $out=deflate_add($c,"sample",ZLIB_SYNC_FLUSH); file_put_contents("result.gz",$out);
