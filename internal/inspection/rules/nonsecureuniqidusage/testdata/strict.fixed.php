<?php
declare(strict_types=1);

$ids = array_map(function ($value = '') { return uniqid((string) $value, true); }, range(1, 3));
$one = call_user_func(function ($value = '') { return uniqid((string) $value, true); }, 5);
