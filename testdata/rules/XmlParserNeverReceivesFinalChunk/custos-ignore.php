<?php
// @custos-ignore XmlParserNeverReceivesFinalChunk

$p=xml_parser_create();xml_parse($p,"<root>");xml_parser_free($p);
