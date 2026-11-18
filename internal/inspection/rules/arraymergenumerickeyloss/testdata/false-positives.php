<?php
array_merge(["a"=>1],["b"=>2]);
array_merge([0=>"a",1=>"b"]);
array_merge($unknown);
array_merge([...$unknown]);
array_merge(...$unknown);
array_merge(["042"=>"a"]);
array_merge([$key=>"a"]);
array_merge(["a"]);
array_merge(['label' => 'x', 0 => 'y'], []);
array_merge(['a', 'k' => 'b', 1 => 'c'], $unknown);
array_merge([0 => 'a', 0 => 'b', 1 => 'c'], $unknown);
array_merge(['a', 'b', 2 => 'c'], ['x' => 1, 0 => 'z']);
array_merge(['a', 0 => 'b', 1 => 'c']);
array_merge([$key => 'a', 1 => 'b', 2 => 'c']);
array_merge([$key => 'a', 'b', 2 => 'c']);
