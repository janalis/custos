<?php
function run($link, string $sql) { return mysqli_query($link, $sql); }
function rows($rs) { return mysqli_num_rows($rs); }

$rs = run($link, "SELECT * FROM t WHERE id = {$_GET['id']}");
$n = rows($rs);
echo "<h5>Count ({$n})</h5>";
echo '<p>' . mysqli_num_rows(mysqli_query($link, "SELECT {$_GET['id']}")) . '</p>';
echo '<p>' . mysqli_affected_rows($link) . ' ' . mysqli_insert_id($link) . '</p>';
echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">mysqli_real_escape_string($link, $_GET['name'])</error> . '</p>';
