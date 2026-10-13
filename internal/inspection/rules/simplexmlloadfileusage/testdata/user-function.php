<?php
namespace Feeds;

function simplexml_load_file($path) { return null; }
function file_get_contents($path) { return ''; }

$a = simplexml_load_file('rss.xml');
$b = \Xml\simplexml_load_file('rss.xml');
$c = <error descr="simplexml_load_file() is affected by PHP bug #62577; load the contents with file_get_contents() and parse them with simplexml_load_string().">\simplexml_load_file('atom.xml')</error>;
