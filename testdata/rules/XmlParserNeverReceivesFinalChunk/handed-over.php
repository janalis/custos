<?php
// Another scope using the same variable name hands nothing over here.
function elsewhere($o) { $p = xml_parser_create(); $o->finish($p); }
$o = new Feed();
$p = xml_parser_create(); xml_parse($p, '<a>'); $o->log(xml_get_error_code($p)); <warning descr="Finish parsing before freeing the XML parser.">xml_parser_free($p)</warning>; $o->finish($p);

class Feed {
    function readsOnly() { $p = xml_parser_create(); xml_parse($p, '<a>'); $this->log(xml_get_current_line_number($p)); Feed::note(strlen('x')); <warning descr="Finish parsing before freeing the XML parser.">xml_parser_free($p)</warning>; }
    function viaMethod() { $p = xml_parser_create(); xml_parse($p, '<a>'); $this->finish($p); xml_parser_free($p); }
    function viaNullsafe(?Feed $o) { $p = xml_parser_create(); xml_parse($p, '<a>'); $o?->finish($p); xml_parser_free($p); }
    function viaStatic() { $p = xml_parser_create(); xml_parse($p, '<a>'); Feed::complete('x', $p); xml_parser_free($p); }
    function viaConstructor() { $p = xml_parser_create(); xml_parse($p, '<a>'); new Closer($p); xml_parser_free($p); }
    function viaFunction() { $p = xml_parser_create(); xml_parse($p, '<a>'); finish_later('x', [$p]); xml_parser_free($p); }
    function viaArrow() { $p = xml_parser_create(); xml_parse($p, '<a>'); $f = fn() => $this->finish($p); $f(); xml_parser_free($p); }
    function finish($p) { xml_parse($p, '</a>', true); }
    function log($v) {}
    static function note($v) {}
    static function complete($label, $p) { xml_parse($p, '</a>', true); }
}
class Closer { function __construct($p) { xml_parse($p, '</a>', true); } }
function finish_later($label, array $parsers) { xml_parse($parsers[0], '</a>', true); }
