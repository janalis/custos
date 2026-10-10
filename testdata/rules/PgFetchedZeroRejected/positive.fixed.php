<?php
function f($r) { if(($v=pg_fetch_result($r,0,0))!==false){echo $v;} }
