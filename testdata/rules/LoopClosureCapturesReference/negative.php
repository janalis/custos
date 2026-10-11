<?php
$jobs=[]; for($i=0;$i<3;$i++){ $jobs[]=function() use ($i){return $i;}; } foreach($jobs as $job){echo $job();}
$jobs=[]; for($i=0;$i<3;$i++){ $jobs[]=function() use (&$i){return function($i){return $i;};}; } foreach($jobs as $job){echo $job();}
$jobs=[]; for($i=0;$i<3;$i++){ $jobs[]=function() use (&$i){$i=1;return 2;}; } foreach($jobs as $job){echo $job();}
$jobs=[]; for($i=0;$i<3;$i++){ $jobs[]=function() use (&$i){[$i]=[10];return 2;}; } foreach($jobs as $job){echo $job();}
$jobs=[]; for($i=0;$i<3;$i++){ $jobs[]=function() use (&$i){list('a'=>list(,$i))=['a'=>[1,2]];return 2;}; } foreach($jobs as $job){echo $job();}
$jobs=[]; for($i=0;$i<3;$i++){ $jobs[]=function() use (&$i){foreach([7] as $i){}return 2;}; } foreach($jobs as $job){echo $job();}
$jobs=[]; for($i=0;$i<3;$i++){ $jobs[]=function() use (&$i){foreach([7] as $i=>$v){}return 2;}; } foreach($jobs as $job){echo $job();}
$jobs=[]; for($i=0;$i<3;$i++){ $jobs[]=function() use (&$i){foreach([[7]] as [$i]){}return 2;}; } foreach($jobs as $job){echo $job();}
$jobs=[]; for($i=0;$i<3;$i++){ $jobs[]=function() use (&$i){unset($i);return 2;}; } foreach($jobs as $job){echo $job();}
