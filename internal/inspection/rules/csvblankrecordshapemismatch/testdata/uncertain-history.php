<?php
$h=fopen('php://memory','w+');if($flag){fwrite($h,"\n");}rewind($h);if(fgetcsv($h)===[]){}
