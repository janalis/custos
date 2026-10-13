<?php
$fp = fopen($path, "r"); if ($fp === false) { return; } <error descr="Open the stream with write access.">fwrite($fp, "change")</error>;
