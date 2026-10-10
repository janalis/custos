<?php
$r = pg_query($db, 'UPDATE jobs SET done = true'); if ($r !== false) { echo <warning descr="Use affected-row count for this statement.">pg_num_rows($r)</warning>; }
