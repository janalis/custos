<?php
$all=[$a,$b]; $write=null; $except=null; while ($running) { $read=$all; stream_select($read,$write,$except,1); }
