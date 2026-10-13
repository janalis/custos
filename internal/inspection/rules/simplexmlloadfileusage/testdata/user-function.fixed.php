<?php
namespace Feeds;

function simplexml_load_file($path) { return null; }
function file_get_contents($path) { return ''; }

$a = simplexml_load_file('rss.xml');
$b = \Xml\simplexml_load_file('rss.xml');
$c = simplexml_load_string(\file_get_contents('atom.xml'));
