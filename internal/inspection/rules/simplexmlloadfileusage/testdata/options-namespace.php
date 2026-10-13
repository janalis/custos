<?php
namespace Feeds;

const LIBXML_NOCDATA = 2;

// The namespace constant wins over the global one: its value is unknown.
$a = <error descr="simplexml_load_file() is affected by PHP bug #62577; load the contents with file_get_contents() and parse them with simplexml_load_string().">simplexml_load_file($p, null, LIBXML_NOCDATA)</error>;
$b = <error descr="simplexml_load_file() is affected by PHP bug #62577; load the contents with file_get_contents() and parse them with simplexml_load_string().">simplexml_load_file($p, null, LIBXML_NOBLANKS)</error>;
