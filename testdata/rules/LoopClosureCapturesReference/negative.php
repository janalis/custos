<?php
$jobs=[]; for($i=0;$i<3;$i++){ $jobs[]=function() use ($i){return $i;}; } foreach($jobs as $job){echo $job();}
$jobs=[]; for($i=0;$i<3;$i++){ $jobs[]=function() use (&$i){return function($i){return $i;};}; } foreach($jobs as $job){echo $job();}
$jobs=[]; for($i=0;$i<3;$i++){ $jobs[]=function() use (&$i){$i=1;return 2;}; } foreach($jobs as $job){echo $job();}
