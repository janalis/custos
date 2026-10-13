<?php
use function gzencode as builtinCall2;
use function gzuncompress as builtinCall6;
<warning descr="Match the decoder to the compression format.">builtinCall6(builtinCall2("sample payload"))</warning>;
