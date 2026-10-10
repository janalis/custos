<?php
$header = fread($stream, 4); if (strlen($header) !== 4) { return; } unpack('Nsize', $header);
