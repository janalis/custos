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
