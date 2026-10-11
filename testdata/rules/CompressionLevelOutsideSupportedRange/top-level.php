<?php

$topLevel = 12;
$packed = <error descr="Use a supported compression level.">gzencode(file_get_contents(__FILE__), $topLevel)</error>;
