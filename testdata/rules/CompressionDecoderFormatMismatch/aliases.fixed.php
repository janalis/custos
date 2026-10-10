<?php
use function gzencode as builtinCall2;
use function gzuncompress as builtinCall6;
gzdecode(builtinCall2("sample payload"));
