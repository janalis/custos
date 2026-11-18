<?php
array_map(function($v){return trim($v);},$names);
array_map(fn($v)=>trim($v),$names);
array_map(function($v){echo $v;},$names);
array_map(function($v){trim($v);return null;},$names);
array_map(function($v){trim($v);yield $v;},$names);
array_map(function($v){foo($v);},$names);
array_map($callback,$names);

array_map(function($s):void{trim($s);},$names);

array_map(function($s){if(false){trim($s);}},$names);

array_map(function ($v) use (&$total) { str_replace('x', 'y', $v, $n); $total += $n; }, $names);
array_map(function ($v) use (&$total) { str_replace('x', 'y', $v, $total); }, $names);
array_map(function ($v) use (&$total) { str_replace('x', 'y', $v, count: $total); }, $names);
array_map(function ($v) use ($args) { str_replace(...$args); }, $names);
