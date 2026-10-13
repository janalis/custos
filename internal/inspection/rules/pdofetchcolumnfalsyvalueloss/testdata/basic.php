<?php
function bad(PDOStatement $s) { if (<warning descr="Compare fetch failure strictly with false.">$value = $s->fetchColumn()</warning>) { echo $value; } }
