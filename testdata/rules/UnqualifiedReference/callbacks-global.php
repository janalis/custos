<?php
// No namespace: callbacks and calls already bind to the global functions.
$ids = array_map('intval', $_GET['ids'] ?? []);
$names = array_filter($ids, 'is_int');
echo call_user_func('strtoupper', 'x'), strlen('y');
