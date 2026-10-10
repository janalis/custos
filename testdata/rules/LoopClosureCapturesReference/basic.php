<?php
$jobs=[]; for($i=0;$i<3;$i++){ $jobs[]=<warning descr="Capture the iteration counter by value for deferred callbacks.">function() use (&$i){return $i;}</warning>; } foreach($jobs as $job){echo $job();}
