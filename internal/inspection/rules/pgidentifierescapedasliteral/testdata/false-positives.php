<?php
// The operand after FROM is a value in these shapes, so the literal API is right.
function distinctFrom($c, $nick) { $v = pg_escape_literal($c, $nick); return pg_query($c, "SELECT id FROM members WHERE nickname IS DISTINCT FROM $v"); }
function notDistinctFrom($c, $nick) { return pg_query($c, "SELECT id FROM members WHERE nickname IS NOT DISTINCT FROM " . pg_escape_literal($c, $nick)); }
function copyFromFile($c, $path) { $v = pg_escape_literal($c, $path); return pg_query($c, "COPY stock FROM $v"); }
function substringFrom($c, $pattern) { $v = pg_escape_literal($c, $pattern); return pg_query($c, "SELECT substring(label FROM $v ) FROM stock"); }
function extractFrom($c, $stamp) { $v = pg_escape_literal($c, $stamp); return pg_query($c, "SELECT id FROM stock WHERE extract(year FROM $v ) = 2020"); }
function unbalanced($c, $name) { $v = pg_escape_literal($c, $name); return pg_query($c, "SELECT 1) FROM $v"); }
