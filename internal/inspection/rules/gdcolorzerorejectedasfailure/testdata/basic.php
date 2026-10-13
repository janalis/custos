<?php
if(<warning descr="Compare color allocation failure strictly with false.">!$color=imagecolorallocate($i,0,0,0)</warning>){throw new RuntimeException();}function color($i){if(<warning descr="Compare color allocation failure strictly with false.">imagecolorallocatealpha($i,0,0,0,1)==false</warning>){return false;}}
