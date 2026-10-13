<?php
$parts = parse_url($raw); if ($parts === false) { throw new RuntimeException(); } echo $parts['host'];
