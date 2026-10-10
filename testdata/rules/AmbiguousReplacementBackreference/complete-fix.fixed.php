<?php
$multiple = preg_replace('/(r)/', '${1}1 and ${1}2', 'r');
$double = preg_replace('/(r)/', "\${1}1", 'r');
$unescaped = preg_replace('/(r)/', "\${1}1", 'r');
$quoted = preg_replace('/(r)/', 'it\'s ${1}1\n', 'r');
$escaped = preg_replace('/(r)/', '\\$11 and ${1}2', 'r');
$encoded = preg_replace('/(r)/', "\x2411", 'r');
