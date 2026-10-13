<?php
echo <warning descr="Check decompression before consuming data.">strlen(gzdecode(gzencode("hello", encoding: ZLIB_ENCODING_DEFLATE)))</warning>;
