<?php
$target = $_GET['next']; if (!in_array($target, ['/home', '/help'], true)) { exit; } header('Location: ' . $target);
