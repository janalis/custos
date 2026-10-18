<?php
// A rewrite through a callback is judged by the callback when a literal names a PHP function.
function escaped() { $_GET['a'] = array_map('htmlspecialchars', $_GET['a']); echo '<p>' . $_GET['a'][0]; }
function trimmed() { $_GET['a'] = array_map('trim', $_GET['a']); echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$_GET['a'][0]</error>; }
function qualified() { $_GET['a'] = array_map('\intval', $_GET['a']); echo '<p>' . $_GET['a'][0]; }
function closure() { $_GET['a'] = array_map(fn($s) => htmlspecialchars($s), $_GET['a']); echo '<p>' . $_GET['a'][0]; }
function project() { $_GET['a'] = array_map('App\esc', $_GET['a']); echo '<p>' . $_GET['a'][0]; }
function spread(array $args) { $_GET['a'] = array_map(...$args); echo '<p>' . $_GET['a'][0]; }
