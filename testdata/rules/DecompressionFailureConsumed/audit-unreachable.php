<?php function audit(){return;
echo strlen(gzdecode(gzencode("hello", encoding: ZLIB_ENCODING_DEFLATE)));
}