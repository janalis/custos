<?php
json_encode(array_filter([1, 2, 0]));
json_encode(array_filter([1, $x]));
json_encode(array_filter([2 => 1, 3 => 0]));
json_encode(array_values(array_filter([0,1])));
json_encode(array_filter([0,1], fn($x)=>true));
json_encode($x);

json_encode(array_filter([0,1]),JSON_FORCE_OBJECT);
