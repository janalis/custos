<?php
$c=inflate_init(ZLIB_ENCODING_RAW); <warning descr="Match the inflate encoding to the input.">inflate_add($c,gzencode("sample"))</warning>;
