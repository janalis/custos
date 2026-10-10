<?php
$p=xml_parser_create();xml_parse($p,"<root>");<warning descr="Finish parsing before freeing the XML parser.">xml_parser_free($p)</warning>;
