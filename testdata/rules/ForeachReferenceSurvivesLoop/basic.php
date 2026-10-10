<?php
$xs=[1,2]; foreach($xs as &$x) {} <warning descr="Unset the foreach reference before reusing the variable.">$x=9</warning>;
