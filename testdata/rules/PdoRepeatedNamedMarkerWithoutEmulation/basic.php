<?php
$pdo = new PDO('mysql:host=localhost;dbname=sample'); $pdo->setAttribute(PDO::ATTR_EMULATE_PREPARES, false); <warning descr="Use a distinct marker for each native parameter.">$pdo->prepare('SELECT :n + :n')</warning>;
