<?php
// A method PHP itself declares takes its documented text parameter as text.
function modified($path) { $date = new DateTime(); <warning descr="Handle file reading failure before using its result.">$date->modify(file_get_contents($path))</warning>; }
function loaded($path) { $document = new DOMDocument(); <warning descr="Handle file reading failure before using its result.">$document->loadXML(file_get_contents($path))</warning>; }
function immutable($path) { $date = new DateTimeImmutable(); <warning descr="Handle file reading failure before using its result.">$date->modify(file_get_contents($path))</warning>; }
