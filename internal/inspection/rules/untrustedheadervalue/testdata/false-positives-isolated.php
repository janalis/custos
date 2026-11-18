<?php
$label = $_GET['label']; if (str_contains($label, "\r") || str_contains($label, "\n")) { exit; } header('X-Label: ' . $label);
