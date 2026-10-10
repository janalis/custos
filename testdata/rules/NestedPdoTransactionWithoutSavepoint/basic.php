<?php
$pdo=new PDO("sqlite::memory:");$pdo->beginTransaction();<warning descr="Finish the transaction before beginning another one.">$pdo->beginTransaction()</warning>;
