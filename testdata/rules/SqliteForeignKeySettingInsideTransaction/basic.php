<?php
$db = new SQLite3(':memory:'); <warning descr="Change foreign-key enforcement before the transaction.">$db->exec('BEGIN; PRAGMA foreign_keys=ON;')</warning>;
