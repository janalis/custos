<?php
$xs=[1,2]; foreach($xs as &$x) {} unset($x); $x=9;
