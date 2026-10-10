<?php
$db = new SQLite3(':memory:'); <warning descr="Remove the deterministic flag from this callback.">$db->createFunction('dice', fn() => random_int(1, 6), 0, SQLITE3_DETERMINISTIC)</warning>;

<warning descr="Remove the deterministic flag from this callback.">$db->createFunction('clock', fn() => time(), 0, SQLITE3_DETERMINISTIC)</warning>;
