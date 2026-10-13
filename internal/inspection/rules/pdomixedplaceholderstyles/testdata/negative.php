<?php
function good(PDO $pdo) { $s = $pdo->prepare("SELECT ':ignore', a FROM items WHERE b = :b"); }
