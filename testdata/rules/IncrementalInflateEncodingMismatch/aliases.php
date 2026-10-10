<?php
use function gzencode as builtinCall2;
use function inflate_add as builtinCall8;
$c=inflate_init(ZLIB_ENCODING_RAW); <warning descr="Match the inflate encoding to the input.">builtinCall8($c,builtinCall2("sample"))</warning>;
