<?php
function benign(PDO $pdo) { $id = intval($_GET['id']); $pdo->query('SELECT * FROM items WHERE id = ' . $id); }
