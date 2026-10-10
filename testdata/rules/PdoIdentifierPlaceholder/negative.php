<?php
function good(PDO $pdo) { $pdo->prepare('SELECT * FROM items WHERE id = :id'); }
