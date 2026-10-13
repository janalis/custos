<?php
$c=inflate_init(ZLIB_ENCODING_DEFLATE);inflate_add($c,gzencode("hello", encoding: ZLIB_ENCODING_DEFLATE));
