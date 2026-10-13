<?php
echo gzuncompress(gzencode("hello", encoding: ZLIB_ENCODING_DEFLATE));
echo gzinflate(gzcompress("hello", encoding: ZLIB_ENCODING_RAW));
echo gzdecode(gzdeflate("hello", encoding: ZLIB_ENCODING_GZIP));
