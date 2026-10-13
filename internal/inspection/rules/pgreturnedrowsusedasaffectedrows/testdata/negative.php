<?php
$r = pg_query($db, 'UPDATE jobs SET done = true'); if ($r !== false) { echo pg_affected_rows($r); }
