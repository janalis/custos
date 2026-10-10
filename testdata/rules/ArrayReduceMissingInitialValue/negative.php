<?php
array_reduce($parts, fn($a,$v)=>array_merge($a,[$v]), []);
array_reduce($parts, fn($a,$v)=>$a+$v);
array_reduce($parts, fn($a,$v)=>array_merge($a??[],[$v]));
array_reduce($parts, function($a,$v){$a=[];return array_merge($a,[$v]);});
array_reduce($parts, function($a,$v){array_merge($a,[$v]);});
array_reduce($parts,$callback);
array_reduce($parts,function(){return array_merge([],[]);});
