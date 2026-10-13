<?php
$doc = simplexml_load_string(file_get_contents($path));
$cfg = simplexml_load_string(file_get_contents($path), null, LIBXML_NOCDATA);
