<?php
// The arrow function parameter is a different array from the parsed one.
parse_str("user.name=Ada", $q);
$pick = fn($q) => $q["user.name"];
echo $pick(["user.name" => "row value"]), "\n";

// A reference binding creates the dotted key next to the normalized one.
parse_str("x.y=1", $w);
$slot = &$w["x.y"];
$slot = 'new';

parse_str("x.y=1", $r);
$refs = [&$r["x.y"]];

parse_str("x.y[]=1", $l);
foreach ($l["x.y"] as &$item) {
    $item = 'seen';
}
