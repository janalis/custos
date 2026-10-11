<?php
// From PHP 8.0 an explicit null reads the current name.
session_start();
echo session_name();
echo session_name(null);
echo session_name(name: null);
echo \session_name(NULL);
$current = null;
echo session_name($current);
