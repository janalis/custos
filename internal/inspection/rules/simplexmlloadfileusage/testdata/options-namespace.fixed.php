<?php
namespace Feeds;

const LIBXML_NOCDATA = 2;

// The namespace constant wins over the global one: its value is unknown.
$a = simplexml_load_file($p, null, LIBXML_NOCDATA);
$b = simplexml_load_string(file_get_contents($p), null, LIBXML_NOBLANKS);
