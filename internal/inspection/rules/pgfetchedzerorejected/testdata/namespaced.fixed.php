<?php
namespace App;

function is_string($value) { return true; }
function f($r) { if (\is_string($v = pg_fetch_result($r, 0, 0))) { echo $v; } }
