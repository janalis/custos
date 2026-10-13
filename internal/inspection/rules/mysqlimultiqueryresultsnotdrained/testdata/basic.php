<?php
$db = new mysqli(); if ($db->multi_query('SELECT 7; SELECT 11')) { <warning descr="Drain all multi-query results before another query.">$db->query('SELECT 14')</warning>; }
