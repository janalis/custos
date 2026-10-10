<?php
$p=xml_parser_create();xml_parse($p,"<root/>",true);xml_parser_free($p);$p=xml_parser_create();xml_parser_free($p);$p=xml_parser_create();xml_parse_into_struct($p,"<r/>",$values);xml_parser_free($p);$p=xml_parser_create();if($x){xml_parse($p,"x");}xml_parser_free($p);$unknown=foo();xml_parser_free($unknown);function dead(){return;xml_parser_free($p);}

strlen("unrelated");
$p=xml_parser_create();xml_parse($p,'<r>');$alias=&$p;unknown($alias);xml_parser_free($p);
