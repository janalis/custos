<?php
// A parser assigned on several paths.

function finalChunkOnOnePath($flag)
{
    if ($flag) {
        $p = xml_parser_create();
        xml_parse($p, "<root>");
    } else {
        $p = xml_parser_create();
        xml_parse($p, "<root/>", true);
    }
    xml_parser_free($p);
}

function finishedInTheBranchThatReplacesIt($flag, $other)
{
    $p = xml_parser_create();
    if ($flag) {
        $p = $other;
        xml_parse($p, "</root>", true);
    } else {
        xml_parse($p, "<root>");
    }
    xml_parser_free($p);
}

// Whichever parser is freed, its last chunk was not final.
function otherParserOnOnePath($flag, $other)
{
    if ($flag) {
        $p = xml_parser_create();
    } else {
        $p = $other;
    }
    xml_parse($p, "<root>");
    <warning descr="Finish parsing before freeing the XML parser.">xml_parser_free($p)</warning>;
}

// The created parser is freed unfinished on the path that keeps it.
function replacedOnOnePathAfterParsing($flag, $other)
{
    $p = xml_parser_create();
    xml_parse($p, "<root>");
    if ($flag) {
        $p = $other;
    }
    <warning descr="Finish parsing before freeing the XML parser.">xml_parser_free($p)</warning>;
}
