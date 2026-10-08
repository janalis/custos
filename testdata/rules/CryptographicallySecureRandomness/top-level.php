<?php
// Top-level code counts as checked (no enclosing function).
$top = openssl_random_pseudo_bytes(4, $strength);
