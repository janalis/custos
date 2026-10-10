<?php namespace Independent;
function xml_parser_free(){}

$p=xml_parser_create();xml_parse($p,"<root>");xml_parser_free($p);
