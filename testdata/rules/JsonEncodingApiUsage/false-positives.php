<?php
namespace Custom;

function json_decode($s) { return $s; }

$a = json_decode('{}');
$b = \json_encode([], JSON_THROW_ON_ERROR);
$c = \json_decode('{}', true, 512, JSON_THROW_ON_ERROR);
