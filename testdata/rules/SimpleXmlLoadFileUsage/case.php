<?php
$doc = <error descr="simplexml_load_file() is affected by PHP bug #62577; load the contents with file_get_contents() and parse them with simplexml_load_string().">SimpleXML_Load_File($path)</error>;
$cfg = <error descr="simplexml_load_file() is affected by PHP bug #62577; load the contents with file_get_contents() and parse them with simplexml_load_string().">\SIMPLEXML_LOAD_FILE($path, null, LIBXML_NOCDATA)</error>;
