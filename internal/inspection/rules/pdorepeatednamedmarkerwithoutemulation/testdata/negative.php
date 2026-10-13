<?php
$pdo = new PDO('mysql:host=localhost;dbname=sample'); $pdo->setAttribute(PDO::ATTR_EMULATE_PREPARES, false); $pdo->prepare('SELECT :a + :b');
